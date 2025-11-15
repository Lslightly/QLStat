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

func batchBuild(cfg *config.Artifact) {
	logdir := cfg.PassLogDir("build")
	// Create output files
	csvFilePath := filepath.Join(logdir, "repoTimes.csv")
	csvFile, err := os.Create(csvFilePath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer csvFile.Close()
	csvFile.WriteString("repo,execution_time\n")

	logFilePath := filepath.Join(logdir, fmt.Sprintf("%s_log.txt", cfg.Lang))
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	// Get list of repositories
	var repos []string
	if len(cfg.BuildRepos) == 1 && cfg.BuildRepos[0] == "-" { // build all repositories
		des, err := os.ReadDir(cfg.RepoRoot)
		if err != nil {
			log.Fatalf("Failed to read directory: %s", err)
		}
		for _, de := range des {
			if de.IsDir() {
				repos = append(repos, filepath.Join(cfg.RepoRoot, de.Name()))
			}
		}
	} else {
		for _, repoName := range cfg.BuildRepos {
			repoPath := filepath.Join(cfg.RepoRoot, repoName)
			if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
				repos = append(repos, repoPath)
			} else {
				log.Printf("Repository does not exist: %s", repoName)
			}
		}
	}

	if len(repos) < 10 {
		log.Println("Repositories to process:", repos)
	} else {
		log.Println("Process", len(repos), "repositories")
	}

	bar := progressbar.Default(int64(len(repos)), "Processing")
	defer bar.Close()
	const (
		Succ int = iota
		ExceedDDL
		Fail
	)
	type CreateDBResult struct {
		status   int
		repoName string
		time     string
		output   string
	}
	resChan := make(chan CreateDBResult, 10)
	var wg sync.WaitGroup

	for _, repo := range repos {
		repoName := filepath.Base(repo)
		dbPath := filepath.Join(cfg.DBRoot, repoName)

		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.BuildTimeout)*time.Second)
			defer cancel()

			cmd := exec.CommandContext(ctx, "codeql", "database", "create", dbPath, "-l="+cfg.Lang, "--overwrite", "-s="+repo)

			log.Printf("Executing command: %s", cmd.String())

			startTime := time.Now()
			output, err := cmd.CombinedOutput()
			executionTime := time.Since(startTime)

			if ctx.Err() == context.DeadlineExceeded {
				resChan <- CreateDBResult{
					status:   ExceedDDL,
					repoName: repoName,
					time:     "",
					output:   string(output),
				}
			} else if err != nil {
				resChan <- CreateDBResult{
					status:   Fail,
					repoName: repoName,
					time:     "",
					output:   string(output),
				}
			} else {
				resChan <- CreateDBResult{
					status:   Succ,
					repoName: repoName,
					time:     executionTime.String(),
					output:   "",
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resChan)
	}()

	failedList := make([]string, 0)
	for res := range resChan {
		t := res.time
		logstatus := "success"
		switch res.status {
		case ExceedDDL:
			t = ""
			logstatus = "timeout"
		case Fail:
			t = ""
			logstatus = "fail"
		}
		if res.status != Succ {
			failedList = append(failedList, res.repoName)
		}
		fmt.Fprintf(csvFile, "%s,%s\n", res.repoName, t)
		fmt.Fprintf(logFile, "%s,%s\n", res.repoName, logstatus)
	}

	cleanupScriptPath := filepath.Join(cfg.DBRoot, "cleanup_failed_directories.sh")
	if len(failedList) > 0 {
		var scriptContent strings.Builder
		scriptContent.WriteString(`#!/bin/bash
# This script removes database directories for builds that failed or timed out.
cd "$(dirname "$0")" || exit
failed_list=(
`)
		for _, repoName := range failedList {
			scriptContent.WriteString("\t\"" + repoName + "\"\n")
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

	log.Printf("Script execution finished. Results written to: %s, log written to: %s", csvFilePath, logFilePath)
}
