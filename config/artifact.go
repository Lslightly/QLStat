package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/Lslightly/qlstat/utils"
	"github.com/goccy/go-yaml"
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
	QueryRepos    []string `yaml:"queryRepos"`
	Queries       []string `yaml:"queries"`
	Externals     []string `yaml:"externals"`
	ExternalFiles []string `yaml:"externalFiles"`
}

// ReadExternalFiles reads filename and returns a slice of non-empty external predicates' names defined in the file
func ReadExternalFiles(filename string) (externals []string, err error) {
	var exts []string
	bs, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	yaml.Unmarshal(bs, &exts)
	return exts, nil
}

type ExternalGenGroup struct {
	GenRepos  []string `yaml:"genRepos"`
	GenScript string   `yaml:"genScript"`
}

var Nowstr string = time.Now().Local().Format("0102-150405")

var logDirCache map[string]string = make(map[string]string) // pass to logDir cache

// logMetaTy is the metadata of a log directory
type logMetaTy struct {
	Pass string `json:"pass"`
	Now  string `json:"now"`
}

const logMetaFileName string = "logMeta.json"

func (l *logMetaTy) dump(file string) error {
	bs, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return os.WriteFile(file, bs, 0644)
}
func loadLogMeta(file string) (res logMetaTy, err error) {
	var bs []byte
	bs, err = os.ReadFile(file)
	if err != nil {
		return
	}
	err = json.Unmarshal(bs, &res)
	return
}

// PassLogDir returns the log directory for the pass, logRoot/pass/current
func (art *Artifact) PassLogDir(pass string) string {
	if dir, ok := logDirCache[pass]; ok {
		return dir
	}
	dir := filepath.Join(art.LogRoot, pass, "current")
	if _, err := os.Stat(dir); err != nil {
		utils.MkdirAll(dir)
	}
	logDirCache[pass] = dir
	logMeta := logMetaTy{
		Pass: pass,
		Now:  Nowstr,
	}
	logMeta.dump(filepath.Join(dir, logMetaFileName))
	return dir
}

// ArchiveCurrentIfExist archives the current log directory if it exists according to logMeta.json
func ArchiveCurrentIfExist(art *Artifact, pass string) (exist bool, err error) {
	srcdir := filepath.Join(art.LogRoot, pass, "current")
	if _, err = os.Stat(srcdir); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return true, err
		}
	}
	meta, err := loadLogMeta(filepath.Join(srcdir, logMetaFileName))
	if err != nil {
		return true, err
	}
	tgtdir := filepath.Join(art.LogRoot, pass, "archive", meta.Now)
	utils.MkdirAll(filepath.Dir(tgtdir))
	return true, os.Rename(srcdir, tgtdir)
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
