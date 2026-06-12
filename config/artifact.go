package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Lslightly/qlstat/utils"
	"github.com/goccy/go-yaml"
)

type Artifact struct {
	RepoRoot     string       `yaml:"repoRoot"`
	LogRoot      string       `yaml:"logRoot"`
	Repositories []*RepoGroup `yaml:"repositories"`
	DBRoot       string       `yaml:"dbRoot"`
	Lang         string       `yaml:"language"`
	BuildTimeout int          `yaml:"buildTimeout"`
	BuildGrps    []BuildGroup `yaml:"buildGrps"`
	QueryConfig  `yaml:"queryconfig"`
}

type QueryConfig struct {
	ResultRoot   string       `yaml:"resultRoot"`
	ParallelCore int          `yaml:"parallelCore"`
	QueryGrps    []QueryGroup `yaml:"queryGrps"`
}

type BuildGroup struct {
	BuildRepos   []string `yaml:"buildRepos"`
	DBName       string   `yaml:"dbName"`
	BuildCommand string   `yaml:"buildCmd"`
	ExtGenScript string   `yaml:"extgenScript"`
}

type QueryGroup struct {
	QueryRoot     string   `yaml:"queryRoot"`
	ExternalRoot  string   `yaml:"externalRoot"`
	QueryDBs      []string `yaml:"queryDBs"`
	Queries       []string `yaml:"queries"`
	Externals     []string `yaml:"externals"`
	ExternalFiles []string `yaml:"externalFiles"`
}

// ResolvedQueryRoot returns the absolute path for this group's query root.
// "std" or empty → <projectRoot>/qlsrc (built-in queries).
// Other values are returned as-is (absolute path or CWD-relative).
func (g *QueryGroup) ResolvedQueryRoot() string {
	if g.QueryRoot == "" || g.QueryRoot == "std" {
		return filepath.Join(utils.ProjectRoot(), "qlsrc")
	}
	return g.QueryRoot
}

// ResolvedExternalRoot returns the absolute path for external predicate files.
// "std" or empty → <projectRoot>/qlsrc (same as built-in query root).
// Other values are returned as-is (absolute path or CWD-relative).
func (g *QueryGroup) ResolvedExternalRoot() string {
	if g.ExternalRoot == "" || g.ExternalRoot == "std" {
		return filepath.Join(utils.ProjectRoot(), "qlsrc")
	}
	return g.ExternalRoot
}

func UnmarshalArtifact(filename string) *Artifact {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	cfg := new(Artifact)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		log.Fatalf("Failed to parse YAML: %v", err)
	}
	if cfg.LogRoot == "" {
		cfg.LogRoot = filepath.Join(utils.ProjectRoot(), "logs")
		log.Println("\033[33mWARNING: logRoot is empty, set to default value:", cfg.LogRoot, "\033[0m")
	}
	if cfg.RepoRoot == "" {
		cfg.RepoRoot = filepath.Join(utils.ProjectRoot(), "repos")
		log.Println("\033[33mWARNING: repoRoot is empty, set to default value:", cfg.RepoRoot, "\033[0m")
	}
	if cfg.DBRoot == "" {
		cfg.DBRoot = filepath.Join(utils.ProjectRoot(), "codeql-db")
		log.Println("\033[33mWARNING: dbRoot is empty, set to default value:", cfg.DBRoot, "\033[0m")
	}
	if cfg.ResultRoot == "" {
		cfg.ResultRoot = filepath.Join(utils.ProjectRoot(), "codeqlResult")
		log.Println("\033[33mWARNING: resultRoot is empty, set to default value:", cfg.ResultRoot, "\033[0m")
	}
	return cfg
}

// ReadExtsFromFile reads filename and returns a slice of non-empty external predicates' names defined in the file
func ReadExtsFromFile(filename string) (externals []string, err error) {
	var exts []string
	bs, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	yaml.Unmarshal(bs, &exts)
	return exts, nil
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
	buildWrittenInRepositories
	buildAll
)

func getReposType(repos []string) reposType {
	if len(repos) == 1 {
		repo0 := repos[0]
		switch repo0 {
		case "-":
			return buildWrittenInRepositories
		case "*":
			return buildAll
		}
	}
	return buildSpecific
}

func (art *Artifact) ConvStrSliceToRepoSlice(repos []string) (res []Repo) {
	switch getReposType(repos) {
	case buildWrittenInRepositories:
		for _, rg := range art.Repositories {
			res = append(res, rg.GetRepos()...)
		}
		return
	case buildAll:
		for _, rg := range art.Repositories {
			res = append(res, rg.reposInDir(art.RepoRoot)...)
		}
		return
	default:
		for _, rg := range art.Repositories {
			rg.calcRepoCache()
		}
		for _, fullname := range repos {
			for _, rg := range art.Repositories {
				if repo, ok := rg.repoCache[fullname]; ok {
					res = append(res, repo)
				}
			}
		}
		return
	}
}

func (art *Artifact) ConvStrSliceToDBSlice(dbnames []string) (res []DB) {
	for _, dbname := range dbnames {
		res = append(res, DB{
			root: art.DBRoot,
			Name: dbname,
		})
	}
	return
}
