package config

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type GitSource struct {
	Prefix             string   `yaml:"prefix"`
	FullNames          []string `yaml:"fullnames"`
	fullName2RepoCache map[string]Repo
}

func (gs *GitSource) calcFullName2RepoCache() {
	if gs.fullName2RepoCache != nil {
		return
	}
	gs.fullName2RepoCache = make(map[string]Repo)
	baseNameCnt := make(map[string]int)
	for _, fullName := range gs.FullNames {
		dirName := filepath.Base(fullName)
		if count, ok := baseNameCnt[dirName]; ok {
			baseNameCnt[dirName] = count + 1
			dirName += strconv.Itoa(count) // change dirName from basename to basename<count>
		} else {
			baseNameCnt[dirName] = 1 // update baseNameCnt
		}
		repo := Repo{
			FullName:    fullName,
			DirBaseName: dirName,
			GitSource:   gs,
		}
		gs.fullName2RepoCache[fullName] = repo
	}
}

func (gs *GitSource) GetRepos() (res []Repo) {
	if gs.fullName2RepoCache == nil {
		gs.calcFullName2RepoCache()
	}
	for _, repo := range gs.fullName2RepoCache {
		res = append(res, repo)
	}
	return
}

func (gs *GitSource) HostDir(root string) string {
	hostName := gs.hostName()
	return filepath.Join(root, hostName)
}

func (gs *GitSource) CreateRepoRootDir(root string) {
	hostDir := gs.HostDir(root)
	if err := os.MkdirAll(hostDir, 0755); err != nil {
		log.Fatalf("error occurs when mkdir %s\n%v", hostDir, err)
	}
}

func (gs *GitSource) hostName() string {
	u, err := url.Parse(gs.Prefix)
	if err != nil {
		log.Fatalf("error when parsing %s with err %v", gs.Prefix, err)
	}
	return u.Hostname()
}

func (gs *GitSource) reposInDir(root string) (res []Repo) {
	des, err := os.ReadDir(gs.HostDir(root))
	if err != nil {
		log.Fatalf("Failed to read directory: %s", err)
	}
	for _, de := range des {
		if de.IsDir() {
			res = append(res, Repo{
				FullName:    "unknown/" + de.Name(),
				DirBaseName: de.Name(),
				GitSource:   gs,
			})
		}
	}
	return
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

func (art *Artifact) GetQueryRepos() (res []Repo) {
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
		for _, fullname := range art.QueryRepos {
			for _, gs := range art.Sources {
				if repo, ok := gs.fullName2RepoCache[fullname]; ok {
					res = append(res, repo)
				}
			}
		}
		return
	}
}
