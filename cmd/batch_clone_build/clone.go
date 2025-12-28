package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/Lslightly/qlstat/config"
	"github.com/schollz/progressbar/v3"
)

/*
batch_clone clone repositories specified by target.yaml in specified root directory.

Notice that org/repo_name repository is cloned in root/repo_name directory.

If some repositories have same base name, it will automatically rename the repository with suffix and create a conflict map in conflict.txt next to target.yaml
*/

func clone(url, dir string) error {
	cmd := exec.Command("git", "clone", url, dir)
	return cmd.Run()
}

func dirSetup(cfg *config.Artifact) {
	if err := os.MkdirAll(cfg.RepoRoot, 0755); err != nil {
		log.Fatalf("Failed to create clone directory: %v", err)
	}
	if err := os.MkdirAll(cfg.DBRoot, 0755); err != nil {
		log.Fatalf("Failed to create database root directory: %v", err)
	}
}

func batchClone(cfg *config.Artifact) {
	type cloneStatus struct {
		fullname string
		err      error // nil means success
	}

	statusChan := make(chan cloneStatus)
	repoCnt := 0
	var wg sync.WaitGroup
	for _, gs := range cfg.Sources {
		gs.CreateRepoRootDir(cfg.RepoRoot)
		for _, repo := range gs.GetRepos() {
			wg.Add(1)
			repoCnt++
			go func() {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						statusChan <- cloneStatus{
							fullname: repo.FullName,
							err:      fmt.Errorf("unknown recovered error: %v", r),
						}
					}
				}()
				if err := repo.Clone(cfg.RepoRoot); err != nil {
					statusChan <- cloneStatus{
						fullname: repo.FullName,
						err:      err,
					}
					return
				}
				statusChan <- cloneStatus{
					fullname: repo.FullName,
					err:      nil,
				}
			}()
		}
	}

	go func() {
		wg.Wait()
		close(statusChan)
	}()

	bar := progressbar.Default(int64(repoCnt), "Cloning Repositories...")

	fails := make([]cloneStatus, 0)
	for status := range statusChan {
		bar.Add(1)
		if status.err != nil {
			fails = append(fails, status)
		}
	}

	if len(fails) != 0 {
		logdir := cfg.PassLogDir("clone")
		failFile := bypass(os.Create(filepath.Join(logdir, "fail.log")))
		defer failFile.Close()

		for _, fail := range fails {
			fmt.Fprintf(failFile, "%s, %v\n", fail.fullname, fail.err)
		}
	}
}
