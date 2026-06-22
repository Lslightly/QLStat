// Copyright 2026 Qingwei Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
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
	} {
		if dir == "" {
			continue
		}
		utils.MkdirAll(dir)
	}
}

func batchClone(cfg *config.Artifact) {
	const passname = "clone"
	if _, err := config.ArchiveCurrentIfExist(cfg, passname); err != nil {
		log.Fatalf("Failed to archive current log dir: %v", err)
	}
	logdir := cfg.PassLogDir(passname)
	type cloneStatus struct {
		fullname string
		err      error // nil means success
	}

	statusChan := make(chan cloneStatus)
	repoCnt := 0
	var wg sync.WaitGroup
	for _, rg := range cfg.Repositories {
		rg.CreateRepoRootDir(cfg.RepoRoot)
		for _, repo := range rg.GetRepos() {
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
		failFile := utils.CreateFile(filepath.Join(logdir, "fail.log"))
		defer failFile.Close()

		for _, fail := range fails {
			fmt.Fprintf(failFile, "%s, %v\n", fail.fullname, fail.err)
		}
	}
}
