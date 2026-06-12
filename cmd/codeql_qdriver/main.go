package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Lslightly/qlstat/config"
	"github.com/Lslightly/qlstat/utils"
	"github.com/schollz/progressbar/v3"
)

var validFmts = []string{
	"text",
	"csv",
	"json",
	"bqrs",
}

var (
	targetDecodeFmt string
	onlyDecode      bool
	doCollect       bool
)

func init() {
	flag.StringVar(&targetDecodeFmt, "format", "csv", "target format of decode, including text, csv, json, bqrs")
	flag.BoolVar(&onlyDecode, "decode-only", false, "only decoding bqrs files")
	flag.BoolVar(&doCollect, "collect", false, "collect all csv results in one csv. The option takes effect only when format is csv.")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "run queries in parallel\nUsage: go run cmd/codeql_qdriver [options] <config file>")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	cfg := config.UnmarshalArtifact(flag.Arg(0))
	if cfg.ParallelCore != 0 {
		runtime.GOMAXPROCS(cfg.ParallelCore)
	}
	if !onlyDecode {
		if _, err := config.ArchiveCurrentIfExist(cfg, "query"); err != nil {
			log.Fatalf("Failed to archive current log dir: %v", err)
		}
		for grpi, grp := range cfg.QueryGrps {
			fmt.Printf("Grp %d: Executing queries\n", grpi)
			queriesExec(cfg, grp, grp.ResolvedQueryRoot(), targetDecodeFmt)
		}
	} else {
		fmt.Println("Decoding results")
		if _, err := config.ArchiveCurrentIfExist(cfg, "decode"); err != nil {
			log.Fatalf("Failed to archive current log dir: %v", err)
		}
		decodeResults(cfg, targetDecodeFmt)
	}
	if doCollect && targetDecodeFmt == "csv" {
		fmt.Println("Collecting CSVs")
		collectCSVs(cfg)
	}
}

// -- External args helper ------------------------------------------------

// buildGroupExternalArgs collects all --external=... options across all queries.
// Since externals are defined at the QueryGroup level, all queries share them,
// but we deduplicate to be safe.
func buildGroupExternalArgs(queries []config.Query, extroot string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, q := range queries {
		for _, ext := range q.ExternalOptions(extroot) {
			if !seen[ext] {
				seen[ext] = true
				result = append(result, ext)
			}
		}
	}
	return result
}

// -- Query execution -----------------------------------------------------

// runQueriesOnDB runs all queries on a single database using codeql database run-queries,
// then decodes the resulting bqrs files to the target format.
//
// Command:
//
//	codeql database run-queries --search-path=... --external=... -- <dbPath> <query1> <query2> ...
//
// Results are placed at <dbPath>/results/<queryPack>/<queryPathNoExt>.bqrs.
// Decoded output goes to <dbPath>/results/<queryPack>/<queryPathNoExt>.<tgtFmt>.
func runQueriesOnDB(cfg *config.Artifact, db config.DB, packName, queryRoot string, queries []config.Query, tgtFmt string) {
	// --- Phase 1: run-queries ---
	extArgs := buildGroupExternalArgs(queries, db.ExtDir())

	queryPaths := make([]string, len(queries))
	for i, q := range queries {
		queryPaths[i] = q.QueryPath(queryRoot)
	}

	logDir := filepath.Join(cfg.PassLogDir("query"), db.Name)
	utils.MkdirAll(logDir)
	outFile, errFile := utils.CreateOutAndErr(filepath.Join(logDir, "run-queries"))
	defer outFile.Close()
	defer errFile.Close()

	args := []string{
		"database", "run-queries",
		"--search-path=" + filepath.Join(utils.ProjectRoot(), "qlsrc/lib"),
	}
	args = append(args, extArgs...)
	args = append(args, "--")
	args = append(args, db.Path())
	args = append(args, queryPaths...)

	cmd := exec.Command("codeql", args...)
	cmd.Stdout = outFile
	cmd.Stderr = errFile
	outFile.WriteString(cmd.String() + "\n")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(errFile, "codeql database run-queries failed: %v\n", err)
	}

	// --- Phase 2: decode bqrs to target format ---
	decodeLogDir := filepath.Join(cfg.PassLogDir("decode"), db.Name)
	utils.MkdirAll(decodeLogDir)

	for _, query := range queries {
		bqrsPath := filepath.Join(db.Path(), "results", packName, query.PathNoExt()+".bqrs")
		csvPath := filepath.Join(db.Path(), "results", packName, query.PathNoExt()+"."+tgtFmt)

		decOut, decErr := utils.CreateOutAndErr(filepath.Join(decodeLogDir, query.Name()))
		cmd := exec.Command("codeql", "bqrs", "decode",
			"--format="+tgtFmt,
			bqrsPath,
			"--output="+csvPath)
		cmd.Stdout = decOut
		cmd.Stderr = decErr
		decOut.WriteString(cmd.String() + "\n")
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(decErr, "codeql bqrs decode failed: %v\n", err)
		}
		decOut.Close()
		decErr.Close()
	}
}

/*
first remove all content in ${config.OutResultRoot}/${qScriptNameNoExt}
dump error log for ${db} in ${config.OutResultRoot}/${qScriptNameNoExt}/error.log
dump stdout/stderr for ${db} in ${config.OutResultRoot}/${qScriptNameNoExt}/log/${db}.log
*/
func queriesExec(cfg *config.Artifact, grp config.QueryGroup, queryRoot string, tgtFmt string) {
	dbs := cfg.ConvStrSliceToDBSlice(grp.QueryDBs)
	packName := resolvePackName(grp)

	// Create Query objects for all queries in the group
	queries := make([]config.Query, len(grp.Queries))
	for i, qScript := range grp.Queries {
		queries[i] = config.CreateQuery(qScript, grp.Externals, grp.ExternalFiles, grp.ResolvedExternalRoot())
	}

	bar := progressbar.Default(int64(len(dbs)))
	var wg sync.WaitGroup

	for _, db := range dbs {
		wg.Add(1)
		go func(db config.DB) {
			defer wg.Done()
			runQueriesOnDB(cfg, db, packName, queryRoot, queries, tgtFmt)
			bar.Describe(fmt.Sprintf("db: %-20s %d queries", db.Name, len(queries)))
			bar.Add(1)
		}(db)
	}

	wg.Wait()
	bar.Close()
}

// -- Decoding ------------------------------------------------------------

func checkDecodeTargetFmt(tgtFmt string) bool {
	if slices.Contains(validFmts, tgtFmt) {
		return true
	}
	log.Fatalln(tgtFmt, "is not valid format, use --help for more information.")
	return false
}

/*
codeql bqrs decode --format=${tgtFmt} ${dbPath}/results/${queryPack}/${queryPathNoExt}.bqrs --output=${dbPath}/results/${queryPack}/${queryPathNoExt}.${tgtFmt}
*/
func decodeResults(cfg *config.Artifact, tgtFmt string) {
	checkDecodeTargetFmt(tgtFmt)

	for grpi, grp := range cfg.QueryGrps {
		dbs := cfg.ConvStrSliceToDBSlice(grp.QueryDBs)
		queries := grp.Queries

		packName := resolvePackName(grp)

		bar := progressbar.Default(int64(len(queries) * len(dbs)))
		for _, qScript := range queries {
			query := config.CreateQuery(qScript, grp.Externals, grp.ExternalFiles, grp.ResolvedExternalRoot())
			for _, db := range dbs {
				bqrsPath := filepath.Join(db.Path(), "results", packName, query.PathNoExt()+".bqrs")
				csvPath := filepath.Join(db.Path(), "results", packName, query.PathNoExt()+"."+tgtFmt)

				decodeLogDir := filepath.Join(cfg.PassLogDir("decode"), db.Name)
				utils.MkdirAll(decodeLogDir)
				decOut, decErr := utils.CreateOutAndErr(filepath.Join(decodeLogDir, query.Name()))

				cmd := exec.Command("codeql", "bqrs", "decode",
					"--format="+tgtFmt,
					bqrsPath,
					"--output="+csvPath)
				cmd.Stdout = decOut
				cmd.Stderr = decErr
				decOut.WriteString(cmd.String() + "\n")
				_ = cmd.Run()
				decOut.Close()
				decErr.Close()

				bar.Describe(fmt.Sprintf("Grp %d: decoding %s/%s", grpi, query.Name(), db.Name))
				bar.Add(1)
			}
		}
		bar.Close()
	}
}

// -- CSV collection ------------------------------------------------------

// resolvePackName resolves the CodeQL pack name for a query group.
// "std" or empty QueryRoot → built-in pack "lslightly/qlstat".
// Otherwise runs "codeql resolve queries" on the first query to get the pack name.
// Only one .ql file per group is needed — same QueryRoot → same qlpack.
func resolvePackName(grp config.QueryGroup) string {
	if grp.QueryRoot == "" || grp.QueryRoot == "std" {
		return "lslightly/qlstat"
	}
	if len(grp.Queries) == 0 {
		return "lslightly/qlstat"
	}
	qPath := filepath.Join(grp.ResolvedQueryRoot(), grp.Queries[0])
	var errBuf bytes.Buffer
	cmd := exec.Command("codeql", "resolve", "queries", qPath)
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to resolve pack for query %s: %v\nOutput: %s", qPath, err, errBuf.String())
	}
	// Parse: "Recording pack reference <pack-name> at <path>"
	line := strings.TrimSpace(errBuf.String())
	prefix := "Recording pack reference "
	rest := strings.TrimPrefix(line, prefix)
	before, _, found := strings.Cut(rest, " at ")
	if !found {
		log.Fatalf("Unexpected codeql resolve output format for %s: %s", qPath, rest)
	}
	return before
}

func collectCSVs(cfg *config.Artifact) {
	for grpi, grp := range cfg.QueryGrps {
		packName := resolvePackName(grp)
		bar := progressbar.Default(int64(len(grp.Queries)))
		for _, qScript := range grp.Queries {
			query := config.CreateQuery(qScript, grp.Externals, grp.ExternalFiles, grp.ResolvedExternalRoot())
			bar.Describe(fmt.Sprintf("Grp %d: Collecting for "+query.PathNoExt(), grpi))
			collectCSVsForQuery(cfg, query, grp, packName)
			bar.Add(1)
		}
	}
}

func collectCSVsForQuery(cfg *config.Artifact, query config.Query, grp config.QueryGroup, packName string) {
	qResultCSV := query.DirPath(cfg.ResultRoot) + ".csv"

	outFile := utils.CreateFile(qResultCSV)
	defer outFile.Close()

	dbs := cfg.ConvStrSliceToDBSlice(grp.QueryDBs)
	var ticket atomic.Int64
	var headerWritten atomic.Bool
	var expectedHeader string
	contentChan := make(chan string, cfg.ParallelCore)
	var hasFile bool

	var wg sync.WaitGroup
	for _, db := range dbs {
		csvPath := filepath.Join(db.Path(), "results", packName, query.PathNoExt()+".csv")

		// Check if the csv file exists
		if _, err := os.Stat(csvPath); os.IsNotExist(err) {
			continue
		}
		hasFile = true

		wg.Add(1)
		go func(csvPath, dbName string) {
			defer wg.Done()

			inFile, err := os.Open(csvPath)
			if err != nil {
				log.Fatalf("Error opening file %s: %v", csvPath, err)
			}
			defer inFile.Close()

			scanner := bufio.NewScanner(inFile)

			if !scanner.Scan() {
				log.Fatalf("Error reading header from %s", csvPath)
			}
			currentHeader := scanner.Text()

			firstFile := ticket.Add(1) == 1
			if firstFile {
				expectedHeader = currentHeader
				headerWritten.Store(true)
			} else {
				for !headerWritten.Load() {
				}
				if currentHeader != expectedHeader {
					log.Fatalf("Warning: Header mismatch in file %s. Expected: %s, Got: %s", csvPath, expectedHeader, currentHeader)
				}
			}
			lines := make([]string, 0)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" {
					break
				}
				lines = append(lines, line+","+dbName)
			}
			contentChan <- strings.Join(lines, "\n")
		}(csvPath, db.Name)
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
	outFile.WriteString(expectedHeader + ",repo\n")
	for c := range contentChan {
		if c == "" {
			continue
		}
		outFile.WriteString(c + "\n")
	}
}
