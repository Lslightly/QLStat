package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Lslightly/qlstat/config"
	yaml "github.com/goccy/go-yaml"
	"github.com/schollz/progressbar/v3"
)

var validFmts = []string{
	"text",
	"csv",
	"json",
	"bqrs",
}

var cfg config.Artifact

var (
	configPath      string
	targetDecodeFmt string
	onlyDecode      bool
)

func init() {
	flag.StringVar(&configPath, "c", "./go.yaml", "the configuration file")
	flag.StringVar(&targetDecodeFmt, "format", "csv", "target format of decode, including text, csv, json, bqrs")
	flag.BoolVar(&onlyDecode, "only-decode", false, "only decoding bqrs files")
	flag.Usage = flag.PrintDefaults
}

func main() {
	flag.Parse()
	parseConfig()
	if !onlyDecode {
		getAnalyzedRepoNames()
		queriesExec()
	}
	decodeResults(targetDecodeFmt, onlyDecode)
}

func parseConfig() {
	bs, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln("error occurs when reading", configPath, err)
	}
	err = yaml.Unmarshal(bs, &cfg)
	if err != nil {
		log.Fatalln("error occurs when parsing json", err)
	}
}

func getAnalyzedRepoNames() {
	if len(cfg.QueryRepos) == 1 && cfg.QueryRepos[0] == "-" { // should get all repositories in DBRoot
		// RepoFlag为空的情况，获取InRootFlag下的所有一级子目录名称并返回
		var repoNames []string
		entries, err := os.ReadDir(cfg.DBRoot)
		if err != nil {
			fmt.Println("Error:", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				repoNames = append(repoNames, entry.Name())
			}
		}
		cfg.QueryRepos = repoNames
	}
}

func qScriptRelatedInfo(qScript string) (qScriptNameNoExt string, qScriptPath string, qResultDir string) {
	qScriptNameNoExt = strings.TrimSuffix(qScript, path.Ext(qScript))
	qScriptPath = path.Join(cfg.QueryRoot, qScript)
	qResultDir = path.Join(cfg.ResultRoot, qScriptNameNoExt)
	return
}

type ErrorPair struct {
	runErr    string
	decodeErr error
}

type QueryStatus struct {
	repo string
	pair ErrorPair
}

func remakeDir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Fatal("error occurs when deleting dir", dir, err)
	}
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatal("error occurs when creating dir", dir, err)
	}
}

/*
first remove all content in ${config.OutResultRoot}/${qScriptNameNoExt}
dump error log for ${repo} in ${config.OutResultRoot}/${qScriptNameNoExt}/error.log
dump stdout/stderr for ${repo} in ${config.OutResultRoot}/${qScriptNameNoExt}/log/${repo}.log
*/
func queriesExec() {
	bar := progressbar.Default(int64(len(cfg.QueryRepos) * len(cfg.Queries)))
	for _, qScript := range cfg.Queries {
		qScriptNameNoExt, qScriptPath, qResultDir := qScriptRelatedInfo(qScript)
		remakeDir(qResultDir)

		queryLogDir := path.Join(cfg.PassLogDir("query"), qScriptNameNoExt)
		err := os.MkdirAll(queryLogDir, 0775)
		if err != nil {
			log.Fatal("error occurs when creating log dir", queryLogDir, err)
		}

		var repo2err map[string]ErrorPair = make(map[string]ErrorPair)
		var wg sync.WaitGroup
		errPairChan := make(chan []QueryStatus, cfg.ParallelCore)
		for _, repoQ := range splitToQueues(cfg.QueryRepos, cfg.ParallelCore) {
			if len(repoQ) == 0 {
				break
			}
			localQ := make(PathQueue, len(repoQ))
			copy(localQ, repoQ)
			wg.Add(1)
			go func() {
				defer wg.Done()
				queryForOneRepo(localQ, bar, qScript, qScriptPath, qResultDir, errPairChan, queryLogDir)
			}()
		}
		go func() {
			wg.Wait()
			close(errPairChan)
		}()
		for statusBuf := range errPairChan {
			for _, status := range statusBuf {
				repo2err[status.repo] = status.pair
			}
		}
		errFilePath := path.Join(cfg.PassLogDir("query"), "error.log")
		errFile, err := os.Create(errFilePath)
		if err != nil {
			log.Fatal("error occurs when create", errFilePath, err)
		}
		defer errFile.Close()
		for repo, err := range repo2err {
			if err.runErr != "" && err.decodeErr != nil {
				fmt.Fprintln(errFile, repo, err.runErr, err.decodeErr)
			}
		}

		bar.Close()
	}
}

type ErrHdr struct {
	args []interface{}
}

func (hdr *ErrHdr) Prefix(args ...interface{}) {
	hdr.args = args
}

func unwrapErr[T any](args ...interface{}) func(v T, err error) T {
	return func(v T, err error) T {
		if err != nil {
			args = append(args, err)
			log.Fatal(args)
		}
		return v
	}
}

func queryForOneRepo(repos PathQueue, bar *progressbar.ProgressBar, qScript string, qScriptPath string, qResultDir string, errChan chan []QueryStatus, logDir string) {
	localStatus := make([]QueryStatus, 0, len(repos))
	for _, repo := range repos {
		repoOutPath := path.Join(logDir, repo+".out")
		repoOut := unwrapErr[*os.File]("error occurs when creating", repoOutPath)(os.Create(repoOutPath))
		defer repoOut.Close()
		repoErrPath := path.Join(logDir, repo+".err")
		repoErr := unwrapErr[*os.File]("error occurs when creating", repoErrPath)(os.Create(repoErrPath))
		defer repoErr.Close()

		var out, err bytes.Buffer
		decodeErr := queryExec(qScriptPath, path.Join(cfg.DBRoot, repo), path.Join(qResultDir, repo), &out, &err)

		bar.Describe(fmt.Sprintf("%-15s %-15s\t", qScript, repo))
		localStatus = append(localStatus, QueryStatus{
			repo: repo,
			pair: ErrorPair{err.String(), decodeErr},
		},
		)

		fmt.Fprint(repoOut, out.String())
		fmt.Fprint(repoErr, err.String())
		bar.Add(1)
	}
	errChan <- localStatus
}

type PathQueue []string

func splitToQueues[T []E, E any](s T, numOfProcess int) (res []T) {
	if len(s) < numOfProcess*20 {
		return []T{s}
	}
	res = make([]T, numOfProcess)
	num := len(s)
	eachNum := num/numOfProcess + 1
	for i := 0; i < numOfProcess; i++ {
		start := min(i*eachNum, num)
		end := min((i+1)*eachNum, num)
		res[i] = s[start:end]
	}
	return
}

/*
codeql query run -d=${config.InDBRoot}/${repo} ${config.QueryRoot}/${qScript} --output=${qResultDir}/${repo}
*/
func queryExec(qScriptPath string, repoDBPath string, outPathNoExt string, out *bytes.Buffer, err *bytes.Buffer) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "codeql", []string{
		"query",
		"run",
		fmt.Sprintf("-d=%s", repoDBPath),
		fmt.Sprintf("--search-path=%s", filepath.Join(cfg.QueryRoot, "lib")),
		qScriptPath,
		fmt.Sprintf("--output=%s", outPathNoExt+".bqrs"),
	}...)
	cmd.Stdout = out
	cmd.Stderr = err
	out.WriteString(cmd.String() + "\n")
	return cmd.Run()
}

func checkDecodeTargetFmt(tgtFmt string) bool {
	for _, fmt := range validFmts {
		if tgtFmt == fmt {
			return true
		}
	}
	log.Fatalln(tgtFmt, "is not valid format, use --help for more information.")
	return false
}

/*
codeql bqrs decode --format=${tgtFmt} ${path}/${fileBase}.bqrs --output=${path}/${fileBase}.${tgtFmt}
*/
func decodeResults(tgtFmt string, decodeOnly bool) {
	checkDecodeTargetFmt(tgtFmt)

	if decodeOnly {
		// 列举config.ResultRoot下的目录
		rootDir := cfg.ResultRoot

		dirNum := len(qlResultDirsUnderRoot(rootDir))

		bar := progressbar.Default(int64(dirNum))
		defer bar.Close()
		filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() && baseInExcludePaths(path) {
				return filepath.SkipDir
			}
			if d.IsDir() {
				bar.Add(1)
				if err = transFilesInDir(tgtFmt, path); err != nil {
					return err
				}
			}
			return nil
		})
		return
	}

	// decode bqrs only in query result dir
	bar := progressbar.Default(int64(len(cfg.Queries)))
	defer bar.Close()
	for _, qScript := range cfg.Queries {
		bar.Add(1)
		_, _, qResultDir := qScriptRelatedInfo(qScript)
		if err := transFilesInDir(tgtFmt, qResultDir); err != nil {
			fmt.Println(err)
			continue
		}
	}
}

/*
ql result directory under `rootDir` whose base is not log/analyze
*/
func qlResultDirsUnderRoot(rootDir string) (res []string) {
	res = make([]string, 0)
	filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			res = append(res, path)
		}
		return nil
	})
	return res
}

func transFilesInDir(tgtFmt string, path string) error {
	fmt.Println("scanning " + path)
	dirEntries, err := os.ReadDir(path)
	if err != nil {
		return fs.SkipDir
	}
	var wg sync.WaitGroup
	for _, groupEntries := range splitToQueues(dirEntries, cfg.ParallelCore) {
		wg.Add(1)
		localQ := make([]fs.DirEntry, len(groupEntries))
		copy(localQ, groupEntries)
		go func() {
			defer wg.Done()
			transFilesForEntries(localQ, path, tgtFmt)
		}()
	}
	wg.Wait()
	return nil
}

func transFilesForEntries(dirEntries []fs.DirEntry, path string, tgtFmt string) {
	for _, dirEntry := range dirEntries {
		if !dirEntry.IsDir() && filepath.Ext(dirEntry.Name()) == ".bqrs" {
			// 执行命令`codeql bqrs decode --format=${tgtFmt} ${path}/${fileBase}.bqrs --output=${path}/${fileBase}.${tgtFmt}`
			bqrsFile := filepath.Join(path, dirEntry.Name())
			outFile := filepath.Join(path, strings.TrimSuffix(dirEntry.Name(), filepath.Ext(dirEntry.Name()))+"."+tgtFmt)

			cmd := exec.Command("codeql", "bqrs", "decode", "--format="+tgtFmt, bqrsFile, "--output="+outFile)
			err := cmd.Run()
			if err != nil {
				continue
			}
		}
	}
}

var excludePaths []string = []string{
	"log",
	"analyze",
}

func baseInExcludePaths(path string) (excluded bool) {
	base := filepath.Base(path)
	excluded = false
	for _, expected := range excludePaths {
		if base == expected {
			return true
		}
	}
	return
}
