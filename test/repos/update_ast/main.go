package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync/atomic"

	"astdb/analyzer"
	"astdb/db"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/tools/go/packages"
)

type Config struct {
	DatabaseConfig db.DatabaseConfig `json:"databaseConfig"`
	SshConfig      db.SshConfig      `json:"sshConfig"`
	ReposDir       string            `json:"reposDir"`
}

var excludeDir = map[string]bool{
	"test":      true,
	"vendor":    true,
	"_obj":      true,
	"build":     true,
	"bin":       true,
	"testdata":  true,
	"_testdata": true,
}

func get_log_file_name(traversedef bool, traversemem bool, traverseMake bool, traverseMakesliceAndNew bool) string {
	if b2i(traversedef)+b2i(traversemem)+b2i(traverseMake) > 1 {
		log.Panic("Option def, mem, makeAndNew are mutual exclusive.")
	}
	if traversedef {
		return "update_ast-def.log"
	} else if traversemem {
		return "update_ast-mem.log"
	} else if traverseMake {
		return "update_ast-make.log"
	} else if traverseMakesliceAndNew {
		return "update_ast-makesliceAndNew.log"
	} else {
		log.Fatalln("please add log file")
		return ""
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func b2i(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}

func main() {
	debug.SetGCPercent(75)
	// parse flags, -trunc: truncate tables
	trunc := false
	interest := false
	traversedef := false
	traversemem := false
	traverseMake := false
	traverseMakesliceAndNew := false
	lightWeightAnalyze := true
	testingFlag := false
	flag.BoolVar(&trunc, "trunc", false, "truncate tables and then exit")
	flag.BoolVar(&interest, "interest", false, "only parse interested repos")
	flag.BoolVar(&traversedef, "def", false, "only traverse def")
	flag.BoolVar(&traversemem, "mem", false, "only traverse mem")
	flag.BoolVar(&traverseMake, "make", false, "only traverse make with no type checking")
	flag.BoolVar(&traverseMakesliceAndNew, "MakesliceAndNew", false, "only traverse makeslice and new with type checking to get size")
	// flag.BoolVar(&lightWeightAnalyze, "lightWeightAnalyze", true, "only lightweight analyze")
	flag.BoolVar(&testingFlag, "testing", false, "testing small repos")
	flag.Parse()
	LOG_FILE_NAME := get_log_file_name(traversedef, traversemem, traverseMake, traverseMakesliceAndNew)
	if traversedef || traversemem || traverseMakesliceAndNew {
		lightWeightAnalyze = false
	} else if traverseMake {
		lightWeightAnalyze = true
	}
	log.SetOutput(os.Stdout)
	var config Config
	config_file, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Error in opening config", err)
	}
	defer config_file.Close()
	decoder := json.NewDecoder(config_file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error in reading config", err)
	}
	analyzer.RepoDir = config.ReposDir
	db.RepoDir = config.ReposDir

	done_list := make(map[string]bool)
	func() {
		log_file, err := os.Open(LOG_FILE_NAME)
		if err != nil {
			return
		}
		defer log_file.Close()
		log_scanner := bufio.NewScanner(log_file)
		for log_scanner.Scan() {
			done_list[log_scanner.Text()] = true
		}
	}()

	err = db.WithConnection(config.DatabaseConfig, config.SshConfig, false, func(connection *db.Connection) error {
		if trunc {
			fmt.Println("Truncating tables...")
			connection.TruncateTables()
			return nil
		}

		var repos []string = []string{"rclone"} // for test
		if !testingFlag {
			repos, err = connection.ListRepoitories(interest)
			if err != nil {
				return err
			}
		}

		fmt.Println("Preparing...")
		bar := progressbar.Default(int64(len(repos)), "Preparing...")
		var subdirs []string
		for _, repo := range repos {
			filepath.WalkDir(filepath.Join(config.ReposDir, repo), func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					return nil
				}
				if done_list[path] {
					return nil
				}
				if excludeDir[d.Name()] || strings.HasSuffix(d.Name(), "_test") || strings.HasPrefix(d.Name(), ".") {
					return filepath.SkipDir
				}
				matches, err := filepath.Glob(fmt.Sprintf("%s/*.go", path))
				if err != nil || len(matches) == 0 {
					return nil
				}
				subdirs = append(subdirs, path)
				return nil
			})
			bar.Add(1)
		}
		bar.Close()

		nproc := runtime.NumCPU() / 4
		// at most 16 transcations a time
		nproc_traverse := min(nproc/2, 10)
		nproc_parse := nproc - nproc_traverse
		path_chan := make(chan string, 10)
		pkg_chan := make(chan analyzer.Result, 10)
		lightWeightPkg_chan := make(chan analyzer.LightWeightResult, 10)
		out_chan := make(chan string, 10)
		ok_chan := make(chan struct{}, 10)

		// analyze dir
		fmt.Println("\nAnalyzing Dir...")
		bar = progressbar.Default(int64(len(subdirs)))
		go func() {
			for _, path := range subdirs {
				bar.Describe(fmt.Sprintf("Analyzing %s\n", path))
				path_chan <- path
			}
			close(path_chan)
		}()
		if !lightWeightAnalyze {
			for i := 0; i < nproc_parse; i++ {
				var mode packages.LoadMode
				if traverseMakesliceAndNew {
					mode = packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo
				} else {
					mode = packages.NeedName | packages.NeedFiles | packages.NeedImports |
						packages.NeedDeps | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedName
				}
				go analyzer.AnalyzeDir(path_chan, pkg_chan, ok_chan, mode)
			}
			go func() {
				cnt := 0
				for _, ok := <-ok_chan; ok; _, ok = <-ok_chan {
					bar.Add(1)
					cnt++
					if cnt == len(subdirs) {
						bar.Close()
						close(pkg_chan)
						close(ok_chan)
					}
				}
			}()
		} else {
			for i := 0; i < nproc_parse; i++ {
				go analyzer.LightWeightAnalyzeDir(path_chan, lightWeightPkg_chan, ok_chan)
			}
			go func() {
				cnt := 0
				for _, ok := <-ok_chan; ok; _, ok = <-ok_chan {
					bar.Add(1)
					cnt++
					if cnt == len(subdirs) {
						bar.Close()
						close(lightWeightPkg_chan)
						close(ok_chan)
					}
				}
			}()
		}

		log_file, err := os.OpenFile(LOG_FILE_NAME, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			log.Fatal(err)
		}
		defer log_file.Close()

		// analyze stms and exprs
		if !lightWeightAnalyze {
			var cnt atomic.Int32
			for i := 0; i < nproc_traverse; i++ {
				go func() {
					for aresult, ok := <-pkg_chan; ok; aresult, ok = <-pkg_chan {
						if traversedef {
							analyzer.TraverseDef(connection, aresult)
						}
						if traversemem {
							analyzer.TraverseStmtAndExpr(connection, aresult)
						}
						if traverseMakesliceAndNew {
							analyzer.TraverseMakesliceAndNew(connection, aresult)
						}
						out_chan <- aresult.Path
					}
					if cnt.Add(1) == int32(nproc_traverse) {
						close(out_chan)
					}
				}()
			}
		} else {
			var cnt atomic.Int32
			for i := 0; i < nproc_traverse; i++ {
				go func() {
					for aresult, ok := <-lightWeightPkg_chan; ok; aresult, ok = <-lightWeightPkg_chan {
						if traverseMake {
							analyzer.LightWeightTraverseMake(connection, aresult)
						}
						out_chan <- aresult.Path
					}
					if cnt.Add(1) == int32(nproc_traverse) {
						close(out_chan)
					}
				}()
			}
		}
		for done_path, ok := <-out_chan; ok; done_path, ok = <-out_chan {
			_, err := fmt.Fprintln(log_file, done_path)
			if err != nil {
				panic(err)
			}
			err = log_file.Sync()
			if err != nil {
				panic(err)
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal("db connection Error:", err)
	}
}
