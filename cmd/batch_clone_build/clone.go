package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Lslightly/qlstat/config"
	"github.com/Lslightly/qlstat/utils"
	"github.com/schollz/progressbar/v3"
)

func dirSetup(cfg *config.Artifact) {
	for _, dir := range []string{
		cfg.RepoRoot,
		cfg.LogRoot,
		cfg.DBRoot,
		cfg.ResultRoot,
	} {
		utils.MkdirAll(dir)
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
				if _, err := os.Stat(repo.DirPath(cfg.RepoRoot)); err == nil {
					// the repo exists
					if err := repo.Checkout(cfg.RepoRoot); err != nil {
						statusChan <- cloneStatus{
							fullname: repo.FullName,
							err:      err,
						}
						return
					}
				} else {
					// the repo does not exist
					if err := repo.Clone(cfg.RepoRoot); err != nil {
						statusChan <- cloneStatus{
							fullname: repo.FullName,
							err:      err,
						}
						return
					}
					if err := repo.Checkout(cfg.RepoRoot); err != nil {
						statusChan <- cloneStatus{
							fullname: repo.FullName,
							err:      err,
						}
					}
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
		failFile := utils.CreateFile(filepath.Join(logdir, "fail.log"))
		defer failFile.Close()

		for _, fail := range fails {
			fmt.Fprintf(failFile, "%s, %v\n", fail.fullname, fail.err)
		}
	}
}
