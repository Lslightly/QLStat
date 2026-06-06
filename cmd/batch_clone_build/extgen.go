package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Lslightly/qlstat/config"
	"github.com/Lslightly/qlstat/utils"
	"github.com/schollz/progressbar/v3"
)

// generate external predicates predicate
// For repositories in each build group, same extgenScript will be applied in the root directory of repositories
// "goescape" means `go build -a -gcflags=-m=2 ./...`. The stderr will be redirected to $dbRoot/<repo>/extgen/m2.log. Then escape_adapter is used to generate databases

func batchExternalGen(cfg *config.Artifact) {
	for grpi, grp := range cfg.BuildGrps {
		if grp.ExtGenScript == "" {
			continue
		}
		fmt.Printf("Grp %d: Gen External Databases\n", grpi)
		genGrp(cfg, grp)
	}
}

func genGrp(cfg *config.Artifact, grp config.BuildGroup) {
	repos := cfg.ConvStrSliceToRepoSlice(grp.BuildRepos)
	var wg sync.WaitGroup
	bar := progressbar.Default(int64(len(repos)))
	for _, repo := range repos {
		wg.Add(1)
		bar.Describe(fmt.Sprintf("Gen external database for %s\n", repo.DirBaseName))
		go func() {
			defer wg.Done()
			defer bar.Add(1)
			utils.MkdirAll(repo.ExtGenDir(cfg.DBRoot)) // mkdir $dbRoot/<repo>/extgen
			if grp.ExtGenScript == "goescape" {
				gobuildM2(cfg, repo)
				adaptEscape(cfg, repo)
			} else {
				genscript(cfg, repo, grp.ExtGenScript)
			}
		}()
	}
	wg.Wait()
}

func escapeLogPath(cfg *config.Artifact, repo config.Repo) string {
	return filepath.Join(repo.ExtGenDir(cfg.DBRoot), "m2.log")
}

func gobuildM2(cfg *config.Artifact, repo config.Repo) {
	logfile := utils.CreateFile(escapeLogPath(cfg, repo))
	defer logfile.Close()
	cmd := exec.Command(
		"go",
		"build",
		"-a",
		"-gcflags=-m=2",
		"./...",
	)
	cmd.Stderr = logfile
	cmd.Dir = repo.DirPath(cfg.RepoRoot)
	_ = cmd.Run()
}
func adaptEscape(cfg *config.Artifact, repo config.Repo) {
	outFile, errFile := utils.CreateOutAndErr(filepath.Join(repo.ExtGenDir(cfg.DBRoot), "adaptEscape"))
	defer outFile.Close()
	defer errFile.Close()
	cmd := exec.Command(
		"go",
		"run",
		"./cmd/escape_adapter",
		"-dir",
		repo.DBExtDir(cfg.DBRoot),
		"-src",
		repo.DirPath(cfg.RepoRoot),
		"-movedToHeap",
		"-newEscapesToHeap",
		escapeLogPath(cfg, repo),
	)
	cmd.Stdout, cmd.Stderr = outFile, errFile
	_ = cmd.Run()
}

func abspath(path string) string {
	p, err := filepath.Abs(path)
	if err != nil {
		log.Fatalf("Fail to get absolute path: %v", err)
	}
	return p
}

func genscript(cfg *config.Artifact, repo config.Repo, script string) {
	outFile, errFile := utils.CreateOutAndErr(filepath.Join(repo.ExtGenDir(cfg.DBRoot), "runscript"))
	defer outFile.Close()
	defer errFile.Close()
	elems := strings.Split(script, " ")
	var cmd *exec.Cmd
	if len(elems) == 1 {
		cmd = exec.Command(elems[0])
	} else {
		cmd = exec.Command(elems[0], elems[1:]...)
	}

	envs := []envpair{
		{REPO_DIR, repo.DirPath(cfg.RepoRoot)},
		{OUTPUT_DIR, repo.ExtGenDir(cfg.DBRoot)},
		{PROJROOT, utils.ProjectRoot()},
		{DB_EXT_DIR, repo.DBExtDir(cfg.DBRoot)},
	}
	cmd.Env = append(os.Environ(), genEnv(allAbs(envs))...)
	cmd.Stdout, cmd.Stderr = outFile, errFile
	fmt.Printf("cwd: %s, out: %s, err: %s, cmd: %s\n", cmd.Dir, outFile.Name(), errFile.Name(), cmd.String())
	_ = cmd.Run()
}
