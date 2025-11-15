package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

	baseNameCnt := make(map[string]int)            // count of existing repos with baseName
	conflictOld2NewName := make(map[string]string) // old name of conflict repos -> new name

	resolveRepoName := func(fullname string) string {
		basename := filepath.Base(fullname)
		if count, ok := baseNameCnt[basename]; ok {
			baseNameCnt[basename] = count + 1
			basename = basename + strconv.Itoa(count) // change basename as basename<cnt>
			conflictOld2NewName[fullname] = basename
		} else {
			baseNameCnt[basename] = 1 // update baseNameCnt
		}
		return basename
	}

	statusChan := make(chan cloneStatus)
	var wg sync.WaitGroup
	repoCnt := 0
	for _, src := range cfg.Sources {
		for _, fullname := range src.FullNames {
			basename := resolveRepoName(fullname)
			tgtdir := filepath.Join(cfg.RepoRoot, basename)
			if _, err := os.Stat(tgtdir); err == nil {
				fmt.Printf("Skiping existing repo: %s\n", fullname)
				continue
			}
			tgturl, err := url.JoinPath(src.Prefix, fullname+".git")
			if err != nil {
				log.Fatalf("Fail to know target url: %v", err)
			}
			wg.Add(1)
			repoCnt++
			go func() {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						statusChan <- cloneStatus{
							fullname: fullname,
							err:      fmt.Errorf("unknown recovered error: %v", r),
						}
					}
				}()
				if err := clone(tgturl, tgtdir); err != nil {
					statusChan <- cloneStatus{
						fullname: fullname,
						err:      err,
					}
					return
				}
				statusChan <- cloneStatus{
					fullname: fullname,
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

	if len(conflictOld2NewName) != 0 || len(fails) != 0 {
		logdir := cfg.PassLogDir("clone")
		conflictFile := bypass(os.Create(filepath.Join(logdir, "conflicts.txt")))
		defer conflictFile.Close()
		failFile := bypass(os.Create(filepath.Join(logdir, "fail.log")))
		defer failFile.Close()

		fmt.Fprintf(conflictFile, "conflicts:\n%v\n", conflictOld2NewName)
		for _, fail := range fails {
			fmt.Fprintf(failFile, "%s, %v\n", fail.fullname, fail.err)
		}
	}
}
