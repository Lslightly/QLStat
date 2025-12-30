package config

import (
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Lslightly/qlstat/utils"
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
	utils.MkdirAll(hostDir)
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
