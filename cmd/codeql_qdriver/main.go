package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
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
	doCollect       bool
)

func init() {
	flag.StringVar(&configPath, "c", "./go.yaml", "the configuration file")
	flag.StringVar(&targetDecodeFmt, "format", "csv", "target format of decode, including text, csv, json, bqrs")
	flag.BoolVar(&onlyDecode, "decode-only", false, "only decoding bqrs files")
	flag.BoolVar(&doCollect, "collect", false, "collect all csv results in one csv. The option takes effect only when format is csv.")
	flag.Usage = flag.PrintDefaults
}

func main() {
	flag.Parse()
	parseConfig()
	if cfg.ParallelCore != 0 {
		runtime.GOMAXPROCS(cfg.ParallelCore)
	}
	if !onlyDecode {
		fmt.Println("Executing queries")
		queriesExec(cfg.GetQueryRepos())
	}
	fmt.Println("Decoding results")
	decodeResults(targetDecodeFmt)
	if doCollect && targetDecodeFmt == "csv" {
		fmt.Println("Collecting CSVs")
		collectCSVs()
	}
}

func parseConfig() {
	bs, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln("error occurs when reading config", configPath, err)
	}
	err = yaml.Unmarshal(bs, &cfg)
	if err != nil {
		log.Fatalln("error occurs when parsing json", err)
	}
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
func queriesExec(repos []config.Repo) {
	bar := progressbar.Default(int64(len(repos) * len(cfg.Queries)))
	for _, qScript := range cfg.Queries {
		query := config.CreateQuery(qScript)

		queryLogDir := path.Join(cfg.PassLogDir("query"), query.PathNoExt())
		err := os.MkdirAll(queryLogDir, 0775)
		if err != nil {
			log.Fatal("error occurs when creating log dir", queryLogDir, err)
		}
		remakeDir(query.AbsPathNoExtWithRoot(cfg.ResultRoot))

		var wg sync.WaitGroup
		for _, repo := range repos {
			wg.Add(1)
			go func(repo config.Repo, query config.Query) {
				defer wg.Done()
				queryForOneRepo(repo, query)
				bar.Describe(fmt.Sprintf("%-15s %-15s\t", query.Name(), repo.DirBaseName))
				bar.Add(1)
			}(repo, query)
		}
		/*
			currently multiple queries for one repository is not supported
			`codeql database run-queries` supports multiple queries.
			But work to collect results from codeql-db/<repo>/results is needed
		*/
		wg.Wait()
	}
	bar.Close()
}

func queryRepoLogSetup(query config.Query, repo config.Repo) (outFile, errFile *os.File) {
	noExtPath := filepath.Join(cfg.PassLogDir("query"), query.PathNoExt(), repo.DirBaseName)
	outpath, errpath := noExtPath+".out", noExtPath+".err"
	var err error
	outFile, err = os.Create(outpath)
	if err != nil {
		log.Fatalf("error occurs when creating %s: %v", outpath, err)
	}
	errFile, err = os.Create(errpath)
	if err != nil {
		log.Fatalf("error occurs when creating %s: %v", errpath, err)
	}
	return
}

/*
codeql query run -d=${config.InDBRoot}/${repo} ${config.QueryRoot}/${qScript} --output=${qResultDir}/${repo}
*/
func queryForOneRepo(repo config.Repo, query config.Query) {
	qResultDir := query.AbsPathNoExtWithRoot(cfg.ResultRoot)
	repoOut, repoErr := queryRepoLogSetup(query, repo)
	defer repoOut.Close()
	defer repoErr.Close()

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "codeql", []string{
		"query",
		"run",
		fmt.Sprintf("-d=%s", repo.DBPath(cfg.DBRoot)),
		fmt.Sprintf("--search-path=%s", filepath.Join(cfg.QueryRoot, "lib")),
		query.AbsPathWithRoot(cfg.QueryRoot),
		fmt.Sprintf("--output=%s", filepath.Join(qResultDir, repo.DirBaseName)+".bqrs"),
	}...)
	cmd.Stdout = repoOut
	cmd.Stderr = repoErr
	repoOut.WriteString(cmd.String() + "\n")
	cmd.Run()
	repoOut.Sync()
	repoErr.Sync()
}

func checkDecodeTargetFmt(tgtFmt string) bool {
	if slices.Contains(validFmts, tgtFmt) {
		return true
	}
	log.Fatalln(tgtFmt, "is not valid format, use --help for more information.")
	return false
}

/*
codeql bqrs decode --format=${tgtFmt} ${path}/${fileBase}.bqrs --output=${path}/${fileBase}.${tgtFmt}
*/
func decodeResults(tgtFmt string) {
	checkDecodeTargetFmt(tgtFmt)

	// decode bqrs only in query result dir
	bar := progressbar.Default(int64(len(cfg.Queries)))
	defer bar.Close()
	for _, qScript := range cfg.Queries {
		bar.Add(1)
		query := config.CreateQuery(qScript)
		bar.Describe(fmt.Sprintf("Decoding %s dir", query.Name()))
		decodeFilesInDir(tgtFmt, query)
	}
}

func decodeLogSetup(decodeLogDir string, repoName string) (outFile, errFile *os.File) {
	var err error
	outpath := filepath.Join(decodeLogDir, repoName+".out")
	outFile, err = os.Create(outpath)
	if err != nil {
		log.Fatal("error when creating", outpath, err)
	}
	errpath := filepath.Join(decodeLogDir, repoName+".err")
	errFile, err = os.Create(errpath)
	if err != nil {
		log.Fatal("error when creating", errpath, err)
	}
	return
}

func decodeFilesInDir(tgtFmt string, query config.Query) {
	decodeLogDir := filepath.Join(cfg.PassLogDir("decode"), query.PathNoExt())
	if err := os.MkdirAll(decodeLogDir, 0755); err != nil {
		log.Fatal(err)
	}

	resultRoot := query.AbsPathNoExtWithRoot(cfg.ResultRoot)
	dirEntries, err := os.ReadDir(resultRoot)
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, de := range dirEntries {
		if de.IsDir() || filepath.Ext(de.Name()) != ".bqrs" {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			outFile, errFile := decodeLogSetup(decodeLogDir, de.Name())
			defer outFile.Close()
			defer errFile.Close()
			tgtPath := filepath.Join(resultRoot, strings.TrimSuffix(de.Name(), filepath.Ext(de.Name()))+"."+tgtFmt)
			cmd := exec.Command("codeql", "bqrs", "decode",
				"--format="+tgtFmt,
				filepath.Join(resultRoot, de.Name()),
				"--output="+tgtPath)
			cmd.Run()
		}()
	}
	wg.Wait()
}

func collectCSVs() {
	bar := progressbar.Default(int64(len(cfg.Queries)))
	for _, qScript := range cfg.Queries {
		query := config.CreateQuery(qScript)
		bar.Describe("collecting for " + query.PathNoExt())
		collectCSVsForQuery(query)
		bar.Add(1)
	}
}

func collectCSVsForQuery(query config.Query) {
	qResultDir := query.AbsPathNoExtWithRoot(cfg.ResultRoot)
	qResultCSV := qResultDir + ".csv"

	// Create or truncate the output CSV file
	outFile, err := os.Create(qResultCSV)
	if err != nil {
		log.Fatalf("Error creating output file %s: %v", qResultCSV, err)
	}
	defer outFile.Close()

	// Read all CSV files in the result directory
	files, err := os.ReadDir(qResultDir)
	if err != nil {
		log.Fatalf("Error reading directory %s: %v", qResultDir, err)
	}

	var ticket atomic.Int64
	var headerWritten atomic.Bool
	var expectedHeader string
	contentChan := make(chan string, cfg.ParallelCore)
	var hasFile bool

	var wg sync.WaitGroup
	for _, file := range files {
		if file.IsDir() || filepath.Ext(file.Name()) != ".csv" {
			continue
		}
		hasFile = true

		wg.Add(1)
		go func(file os.DirEntry, qResultDir string) {
			defer wg.Done()
			// Get repo name from filename (without .csv extension)
			repoName := strings.TrimSuffix(file.Name(), ".csv")

			// For the first file, write header with repo_name column
			fpath := filepath.Join(qResultDir, file.Name())
			// Open the CSV file
			inFile, err := os.Open(fpath)
			if err != nil {
				log.Fatalf("Error opening file %s: %v", fpath, err)
			}
			defer inFile.Close()

			scanner := bufio.NewScanner(inFile)

			if !scanner.Scan() {
				log.Fatalf("Error reading header from %s", fpath)
			}
			// Read header
			currentHeader := scanner.Text()

			firstFile := ticket.Add(1) == 1
			if firstFile {
				expectedHeader = currentHeader
				headerWritten.Store(true)
			} else {
				for !headerWritten.Load() {
				}
				if currentHeader != expectedHeader {
					log.Fatalf("Warning: Header mismatch in file %s. Expected: %s, Got: %s", fpath, expectedHeader, currentHeader)
				}
			}
			lines := make([]string, 0)
			for scanner.Scan() {
				lines = append(lines, scanner.Text()+","+repoName)
			}
			contentChan <- strings.Join(lines, "\n")
		}(file, qResultDir)
	}
	go func() {
		wg.Wait()
		close(contentChan)
	}()
	if !hasFile {
		fmt.Printf("query %s does not have results", query.Name())
		return
	}
	for !headerWritten.Load() {
	}
	outFile.WriteString(expectedHeader + "\n")
	for c := range contentChan {
		outFile.WriteString(c + "\n")
	}
}
