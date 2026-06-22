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

package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Lslightly/qlstat/utils"
)

// RepoGroup groups repositories sharing the same URL prefix and/or directory sub-path.
//   - URLPrefix: if set, URLPrefix + repo name forms the clone URL.
//   - Dir:       if set, repos are stored under repoRoot/<Dir>/<repo>.
//     If empty, repos go directly under repoRoot/<repo>.
//   - Repos:     list of repository names in this group.
type RepoGroup struct {
	URLPrefix string   `yaml:"urlPrefix"`
	Dir       string   `yaml:"dir"`
	Repos     []string `yaml:"repos"`

	repoCache map[string]Repo
}

func (rg *RepoGroup) calcRepoCache() {
	if rg.repoCache != nil {
		return
	}
	rg.repoCache = make(map[string]Repo)
	baseNameCnt := make(map[string]int)
	for _, fullName := range rg.Repos {
		branch := ""
		if strings.Contains(fullName, " ") { // extract branch
			elems := strings.Fields(fullName)
			fullName, branch = elems[0], elems[1]
		}
		dirName := filepath.Base(fullName)
		if count, ok := baseNameCnt[dirName]; ok {
			baseNameCnt[dirName] = count + 1
			dirName += strconv.Itoa(count)
		} else {
			baseNameCnt[dirName] = 1
		}
		repo := Repo{
			FullName:    fullName,
			branch:      branch,
			DirBaseName: dirName,
			RepoGroup:   rg,
		}
		rg.repoCache[fullName] = repo
	}
}

func (rg *RepoGroup) GetRepos() (res []Repo) {
	if rg.repoCache == nil {
		rg.calcRepoCache()
	}
	for _, repo := range rg.repoCache {
		res = append(res, repo)
	}
	return
}

// GroupDir returns filepath.Join(root, Dir) if Dir is non-empty, otherwise root.
func (rg *RepoGroup) GroupDir(root string) string {
	if rg.Dir == "" {
		return root
	}
	return filepath.Join(root, rg.Dir)
}

// CreateRepoRootDir creates the directory for this group under repo root.
func (rg *RepoGroup) CreateRepoRootDir(root string) {
	if rg.Dir == "" {
		return
	}
	utils.MkdirAll(filepath.Join(root, rg.Dir))
}

// reposInDir discovers repos already on disk under this group's directory.
func (rg *RepoGroup) reposInDir(root string) (res []Repo) {
	des, err := os.ReadDir(rg.GroupDir(root))
	if err != nil {
		log.Fatalf("Failed to read directory: %s", err)
	}
	for _, de := range des {
		if de.IsDir() {
			res = append(res, Repo{
				FullName:    "unknown/" + de.Name(),
				DirBaseName: de.Name(),
				RepoGroup:   rg,
			})
		}
	}
	return
}
