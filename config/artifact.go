package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Lslightly/qlstat/utils"
)

type Artifact struct {
	RepoRoot     string             `yaml:"repoRoot"`
	LogRoot      string             `yaml:"logRoot"`
	Sources      []*GitSource       `yaml:"sources"`
	DBRoot       string             `yaml:"dbRoot"`
	Lang         string             `yaml:"language"`
	BuildTimeout int                `yaml:"buildTimeout"`
	BuildGrps    []BuildGroup       `yaml:"buildGrps"`
	ExtGenGrps   []ExternalGenGroup `yaml:"externalGenGrps"`
	QueryConfig  `yaml:"queryconfig"`
}

type QueryConfig struct {
	ResultRoot   string       `yaml:"resultRoot"`
	QueryRoot    string       `yaml:"queryRoot"`
	ParallelCore int          `yaml:"parallelCore"`
	QueryGrps    []QueryGroup `yaml:"queryGrps"`
}

type BuildGroup struct {
	BuildRepos   []string `yaml:"buildRepos"`
	BuildCommand string   `yaml:"buildCmd"`
}

type QueryGroup struct {
	QueryRepos []string `yaml:"queryRepos"`
	Queries    []string `yaml:"queries"`
	Externals  []string `yaml:"externals"`
}

type ExternalGenGroup struct {
	GenRepos  []string `yaml:"genRepos"`
	GenScript string   `yaml:"genScript"`
}

var Nowstr string = time.Now().Local().Format("0102-150405")

var logDirMap map[string]string = make(map[string]string)

func (art *Artifact) PassLogDir(pass string) string {
	if dir, ok := logDirMap[pass]; ok {
		return dir
	}
	dir := filepath.Join(art.LogRoot, pass, Nowstr)
	if _, err := os.Stat(dir); err != nil {
		utils.MkdirAll(dir)
	}
	logDirMap[pass] = dir
	return dir
}

type reposType int

const (
	buildSpecific reposType = iota
	buildWrittenInSources
	buildAll
)

func getReposType(repos []string) reposType {
	if len(repos) == 1 {
		repo0 := repos[0]
		switch repo0 {
		case "-":
			return buildWrittenInSources
		case "*":
			return buildAll
		}
	}
	return buildSpecific
}

func (art *Artifact) ConvStrSliceToRepoSlice(repos []string) (res []Repo) {
	switch getReposType(repos) {
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
		for _, fullname := range repos {
			for _, gs := range art.Sources {
				if repo, ok := gs.fullName2RepoCache[fullname]; ok {
					res = append(res, repo)
				}
			}
		}
		return
	}
}
