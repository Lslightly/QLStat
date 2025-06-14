package main

import (
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

	yaml "github.com/goccy/go-yaml"
	"github.com/schollz/progressbar/v3"
)

/*
希望实现的功能：
DONE 自定义要批量分析的仓库列表
DONE 自定义要分析的查询脚本
DONE 解析查询结果 result_parser
DONE 将查询结果进行批量保存 result_parser
DONE 批量解析查询结果 result_parser
*/

type ConfigTy struct {
	InDBRoot      string   `yaml:"InDBRoot"`
	Repos         []string `yaml:"Repos"`
	QueryRoot     string   `yaml:"QueryRoot"`
	Queries       []string `yaml:"Queries"`
	OutResultRoot string   `yaml:"OutResultRoot"`
	ErrorLog      string   `yaml:"ErrorLog"`
	ParallelCore  int      `yaml:"ParallelCore"`
}

var validFmts = []string{
	"text",
	"csv",
	"json",
	"bqrs",
}

var config ConfigTy

var (
	configPath      string
	targetDecodeFmt string
	onlyDecode      bool
)

var globalLock sync.Mutex

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
	err = yaml.Unmarshal(bs, &config)
	if err != nil {
		log.Fatalln("error occurs when parsing json", err)
	}
}

func getAnalyzedRepoNames() {
	if len(config.Repos) == 0 {
		// RepoFlag为空的情况，获取InRootFlag下的所有一级子目录名称并返回
		var repoNames []string
		entries, err := os.ReadDir(config.InDBRoot)
		if err != nil {
			fmt.Println("Error:", err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				repoNames = append(repoNames, entry.Name())
			}
		}
		config.Repos = repoNames
	}
}

func qScriptRelatedInfo(qScript string) (qScriptNameNoExt string, qScriptPath string, qResultDir string) {
	qScriptNameNoExt = strings.TrimSuffix(qScript, path.Ext(qScript))
	qScriptPath = path.Join(config.QueryRoot, qScript)
	qResultDir = path.Join(config.OutResultRoot, qScriptNameNoExt)
	return
}

type ErrorPair struct {
	runErr    error
	decodeErr error
}

/*
first remove all content in ${config.OutResultRoot}/${qScriptNameNoExt}
dump error log for ${repo} in ${config.OutResultRoot}/${qScriptNameNoExt}/error.log
dump stdout/stderr for ${repo} in ${config.OutResultRoot}/${qScriptNameNoExt}/log/${repo}.log
*/
func queriesExec() {
	bar := progressbar.Default(int64(len(config.Repos) * len(config.Queries)))
	for _, qScript := range config.Queries {
		_, qScriptPath, qResultDir := qScriptRelatedInfo(qScript)
		// first remove $qResultDir
		err := os.RemoveAll(qResultDir)
		if err != nil {
			log.Fatal("error occurs when deleting dir", qResultDir, err)
		}

		logDir := path.Join(qResultDir, "log")
		err = os.MkdirAll(logDir, 0775)
		if err != nil {
			log.Fatal("error occurs when creating log dir", logDir, err)
		}

		var repo2err map[string]ErrorPair = make(map[string]ErrorPair)
		var wg sync.WaitGroup
		for _, repoQ := range splitToQueues(config.Repos, config.ParallelCore) {
			if len(repoQ) == 0 {
				break
			}
			localQ := make(PathQueue, len(repoQ))
			copy(localQ, repoQ)
			wg.Add(1)
			go func() {
				defer wg.Done()
				queryForOneRepo(localQ, bar, qScript, qScriptPath, qResultDir, repo2err, logDir)
			}()
		}
		wg.Wait()
		errFilePath := path.Join(qResultDir, config.ErrorLog)
		errFile, err := os.Create(errFilePath)
		if err != nil {
			log.Fatal("error occurs when create", errFilePath, err)
		}
		defer errFile.Close()
		for repo, err := range repo2err {
			if err.runErr != nil && err.decodeErr != nil {
				fmt.Fprintln(errFile, repo, err.runErr, err.decodeErr)
			}
		}

		bar.Close()
	}
}

func queryForOneRepo(repos PathQueue, bar *progressbar.ProgressBar, qScript string, qScriptPath string, qResultDir string, repo2err map[string]ErrorPair, logDir string) {
	for _, repo := range repos {
		out, err, decodeErr := queryExec(qScriptPath, path.Join(config.InDBRoot, repo), path.Join(qResultDir, repo))

		globalLock.Lock()
		bar.Describe(fmt.Sprintf("%-15s %-15s\t", qScript, repo))
		fmt.Println()
		repo2err[repo] = ErrorPair{
			runErr:    err,
			decodeErr: decodeErr,
		}
		globalLock.Unlock()

		repoOutPath := path.Join(logDir, repo+".log")
		repoOut, err := os.Create(repoOutPath)
		if err != nil {
			log.Fatal("error occurs when creating", repoOutPath, err)
		}
		fmt.Fprint(repoOut, out)

		bar.Add(1)
	}
}

type PathQueue []string

func splitToQueues[T []E, E any](s T, numOfProcess int) (res []T) {
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

func min(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

/*
codeql query run -d=${config.InDBRoot}/${repo} ${config.QueryRoot}/${qScript} --output=${qResultDir}/${repo}
*/
func queryExec(qScriptPath string, repoDBPath string, outPathNoExt string) (string, error, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "codeql", []string{
		"query",
		"run",
		fmt.Sprintf("-d=%s", repoDBPath),
		qScriptPath,
		fmt.Sprintf("--output=%s", outPathNoExt+".bqrs"),
	}...)
	globalLock.Lock()
	fmt.Println(cmd.String())
	globalLock.Unlock()
	out, err := cmd.CombinedOutput()

	return string(out), err, nil
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
		// 列举config.OutResultRoot下的目录
		rootDir := config.OutResultRoot

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
	bar := progressbar.Default(int64(len(config.Queries)))
	defer bar.Close()
	for _, qScript := range config.Queries {
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
	for _, groupEntries := range splitToQueues(dirEntries, config.ParallelCore) {
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
			globalLock.Lock()
			fmt.Println(cmd.String())
			globalLock.Unlock()
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
