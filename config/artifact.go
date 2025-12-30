package config

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

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
	ResultRoot   string       `yaml:"resultRoot"`
	QueryRoot    string       `yaml:"queryRoot"`
	ParallelCore int          `yaml:"parallelCore"`
	Grps         []QueryGroup `yaml:"queryGrps"`
}

type QueryGroup struct {
	QueryRepos []string `yaml:"queryRepos"`
	Queries    []string `yaml:"queries"`
	Externals  []string `yaml:"externals"`
}

var Nowstr string = time.Now().Local().Format("0102-150405")

var logDirMap map[string]string = make(map[string]string)

func (art *Artifact) PassLogDir(pass string) string {
	if dir, ok := logDirMap[pass]; ok {
		return dir
	}
	dir := filepath.Join(art.LogRoot, pass, Nowstr)
	if _, err := os.Stat(dir); err != nil {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatal(err)
		}
	}
	logDirMap[pass] = dir
	return dir
}

type buildType int

const (
	buildSpecific buildType = iota
	buildWrittenInSources
	buildAll
)

func (art *Artifact) getBuildType() buildType {
	if len(art.BuildRepos) == 1 {
		repo0 := art.BuildRepos[0]
		switch repo0 {
		case "-":
			return buildWrittenInSources
		case "*":
			return buildAll
		}
	}
	return buildSpecific
}

func (art *Artifact) GetBuildRepos() (res []Repo) {
	switch art.getBuildType() {
	case buildWrittenInSources:
		for _, gs := range art.Sources {
			res = append(res, gs.GetRepos()...)
		}
		return
	case buildAll:
		for _, gs := range art.Sources {
			res = append(res, gs.reposInDir(art.RepoRoot)...)
		}
		return
	default:
		for _, gs := range art.Sources {
			gs.calcFullName2RepoCache()
		}
		for _, fullname := range art.BuildRepos {
			for _, gs := range art.Sources {
				if repo, ok := gs.fullName2RepoCache[fullname]; ok {
					res = append(res, repo)
				}
			}
		}
		return
	}
}

func (art *Artifact) ConvStrSliceToRepoSlice(queryRepos []string) (res []Repo) {
	switch art.getBuildType() {
	case buildWrittenInSources:
		for _, gs := range art.Sources {
			res = append(res, gs.GetRepos()...)
		}
		return
	case buildAll:
		for _, gs := range art.Sources {
			res = append(res, gs.reposInDir(art.DBRoot)...)
		}
		return
	default:
		for _, gs := range art.Sources {
			gs.calcFullName2RepoCache()
		}
		for _, fullname := range queryRepos {
			for _, gs := range art.Sources {
				if repo, ok := gs.fullName2RepoCache[fullname]; ok {
					res = append(res, repo)
				}
			}
		}
		return
	}
}
