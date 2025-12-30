package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Lslightly/qlstat/config"
	"github.com/schollz/progressbar/v3"
)

type buildStatus int

func (bs buildStatus) String() string {
	switch bs {
	case ExceedDDL:
		return "timeout"
	case Fail:
		return "fail"
	case MKdirFail:
		return "mkdirFail"
	default:
		return "success"
	}
}

const (
	Succ buildStatus = iota
	ExceedDDL
	Fail
	MKdirFail
)

type CreateDBResult struct {
	status buildStatus
	dbPath string
	time   string
}

func batchBuild(cfg *config.Artifact) {
	csvFile, logFile := buildDirSetup(cfg)
	defer csvFile.Close()
	defer logFile.Close()

	repos := cfg.GetBuildRepos()
	if len(repos) < 10 {
		repoNames := make([]string, 0)
		for _, repo := range repos {
			repoNames = append(repoNames, repo.FullName)
		}
		log.Println("Create database for", repoNames)
	} else {
		log.Println("Create database for", len(repos), "repositories")
	}

	bar := progressbar.Default(int64(len(repos)), "Creating database")
	defer bar.Close()
	resChan := make(chan CreateDBResult, 10)
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		go func() {
			defer wg.Done()
			build(cfg, repo, resChan)
		}()
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	failedPaths := make([]string, 0)
	for res := range resChan {
		if res.status != Succ {
			failedPaths = append(failedPaths, res.dbPath)
		}
		fmt.Fprintf(csvFile, "%s,%s\n", res.dbPath, res.time)
		fmt.Fprintf(logFile, "%s,%s\n", res.dbPath, res.status)
	}

	removeCleanupScript(cfg)
	if len(failedPaths) > 0 {
		createCleanupScript(cfg, failedPaths)
	}
	log.Printf("Script execution finished. Results written to: %s, log written to: %s", csvFile.Name(), logFile.Name())
}

func buildDirSetup(cfg *config.Artifact) (*os.File, *os.File) {
	logdir := cfg.PassLogDir("build")
	// Create output files
	csvFilePath := filepath.Join(logdir, "repoTimes.csv")
	csvFile, err := os.Create(csvFilePath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer csvFile.Close()
	csvFile.WriteString("repo,execution_time\n")

	logFilePath := filepath.Join(logdir, "repo_build.txt")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer logFile.Close()

	for _, gs := range cfg.Sources {
		hostdir := gs.HostDir(cfg.RepoRoot)
		if err := os.MkdirAll(filepath.Join(hostdir), 0755); err != nil {
			log.Fatalf("error occurs when mkdir %s\n%v", hostdir, err)
		}
	}
	return csvFile, logFile
}

func removeCleanupScript(cfg *config.Artifact) {
	scriptPath := cfg.DBCleanUpScriptPath()
	if _, err := os.Stat(scriptPath); err == nil {
		os.Remove(scriptPath)
	}
}

func createCleanupScript(cfg *config.Artifact, failedPaths []string) {
	cleanupScriptPath := cfg.DBCleanUpScriptPath()
	var scriptContent strings.Builder
	scriptContent.WriteString(`#!/bin/bash
# This script removes database directories for builds that failed or timed out.
cd "$(dirname "$0")" || exit
failed_list=(
`)
	for _, dbPath := range failedPaths {
		scriptContent.WriteString("\t\"" + dbPath + "\"\n")
	}
	scriptContent.WriteString(`)
for dir in "${failed_list[@]}"; do
    echo "Removing failed directory: $dir"
    rm -rf "$dir"
done
`)

	err := os.WriteFile(cleanupScriptPath, []byte(scriptContent.String()), 0755)
	if err != nil {
		log.Fatalf("Failed to create cleanup script: %v", err)
	}
	log.Printf("Cleanup script for failed directories created at: %s", cleanupScriptPath)
}

func buildRepoLogSetup(cfg *config.Artifact, repo config.Repo) (outFile, errFile *os.File) {
	logdir := cfg.PassLogDir("build")
	repoRootDir := repo.DirPath(logdir)
	var err error
	if err = os.MkdirAll(repoRootDir, 0755); err != nil {
		log.Fatalf("Fail to mkdir %s", repoRootDir)
	}
	outpath := filepath.Join(repoRootDir, "out")
	outFile, err = os.Create(outpath)
	if err != nil {
		log.Fatalf("Fail to create %s", outpath)
	}
	errpath := filepath.Join(repoRootDir, "err")
	errFile, err = os.Create(errpath)
	if err != nil {
		log.Fatalf("Fail to create %s", errpath)
	}
	return
}

func build(cfg *config.Artifact, repo config.Repo, resChan chan CreateDBResult) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.BuildTimeout)*time.Second)
	defer cancel()

	outFile, errFile := buildRepoLogSetup(cfg, repo)
	defer outFile.Close()
	defer errFile.Close()
	dbPath := repo.DBPath(cfg.DBRoot)
	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		resChan <- CreateDBResult{
			status: MKdirFail,
			dbPath: dbPath,
			time:   "",
		}
	}
	cmd := exec.CommandContext(ctx, "codeql", "database", "create", dbPath, "-l="+cfg.Lang, "--overwrite", "-s="+repo.DirPath(cfg.RepoRoot))
	cmd.Stdout = outFile
	cmd.Stderr = errFile

	log.Printf("Executing command: %s", cmd.String())

	startTime := time.Now()
	err = cmd.Run()
	executionTime := time.Since(startTime)

	if ctx.Err() == context.DeadlineExceeded {
		resChan <- CreateDBResult{
			status: ExceedDDL,
			dbPath: dbPath,
			time:   "",
		}
	} else if err != nil {
		resChan <- CreateDBResult{
			status: Fail,
			dbPath: dbPath,
			time:   "",
		}
	} else {
		resChan <- CreateDBResult{
			status: Succ,
			dbPath: dbPath,
			time:   executionTime.String(),
		}
	}
}
