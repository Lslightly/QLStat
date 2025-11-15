package config

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type GitSource struct {
	Prefix    string   `yaml:"prefix"`
	FullNames []string `yaml:"fullnames"`
}

type Artifact struct {
	RepoRoot     string      `yaml:"repoRoot"`
	LogRoot      string      `yaml:"logRoot"`
	Sources      []GitSource `yaml:"sources"`
	DBRoot       string      `yaml:"dbRoot"`
	Lang         string      `yaml:"language"`
	BuildTimeout int         `yaml:"buildTimeout"`
	BuildRepos   []string    `yaml:"buildRepos"` // ["-"] indicates all repositories
	QueryConfig  `yaml:"queryconfig"`
}

type QueryConfig struct {
	ResultRoot   string   `yaml:"resultRoot"`
	QueryRepos   []string `yaml:"queryRepos"`
	QueryRoot    string   `yaml:"queryRoot"`
	Queries      []string `yaml:"queries"`
	ParallelCore int      `yaml:"parallelCore"`
}

var Nowstr string = time.Now().Local().Format("0102-030405")

var logdirmap map[string]string = make(map[string]string)

func (art *Artifact) PassLogDir(pass string) string {
	if dir, ok := logdirmap[pass]; ok {
		return dir
	}
	dir := filepath.Join(art.LogRoot, pass, Nowstr)
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal(err)
		}
	}
	logdirmap[pass] = dir
	return dir
}
