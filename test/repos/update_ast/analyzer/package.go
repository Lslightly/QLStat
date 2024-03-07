package analyzer

import (
	"astdb/db"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"strings"

	"golang.org/x/tools/go/packages"
)

type Result struct {
	Path string
	Pkg  *packages.Package
}

type LightWeightResult struct {
	Path   string
	ASTPkg *ast.Package
	fset   *token.FileSet
}

// Set this variable to config.RepoDir before analyzing.
var RepoDir string = "[uninitialized]"

func AnalyzeDir(in <-chan string, out chan<- Result, okCh chan<- struct{}, mode packages.LoadMode) {
	if RepoDir == "[uninitialized]" {
		panic(nil)
	}
	for path, ok := <-in; ok; path, ok = <-in {
		cfg := packages.Config{
			Mode: mode,
			Dir:  path,
		}
		if pkgs, err := packages.Load(&cfg, path); err == nil {
			for _, pkg := range pkgs {
				out <- Result{path, pkg}
			}
		} else {
			log.Printf("Error while analyzing %s: %v", path, err)
		}
		okCh <- struct{}{}
	}
}

func LightWeightAnalyzeDir(in <-chan string, out chan<- LightWeightResult, okCh chan<- struct{}) {
	if RepoDir == "[uninitialized]" {
		panic(nil)
	}
	for path, ok := <-in; ok; path, ok = <-in {
		fset := token.NewFileSet()
		mode := parser.AllErrors
		if pkgs, err := parser.ParseDir(fset, path, func(fi fs.FileInfo) bool {
			return !strings.HasSuffix(fi.Name(), "_test.go")
		}, mode); err == nil {
			for _, pkg := range pkgs {
				out <- LightWeightResult{path, pkg, fset}
			}
		} else {
			log.Printf("Error while analyzing %s: %v", path, err)
		}
		okCh <- struct{}{}
	}
}

func TraverseDef(conn *db.Connection, aresult Result) {
	pkg := aresult.Pkg
	for _, file := range pkg.Syntax {
		visitor := visitorDef{
			func_ids: []int64{},
			file_set: pkg.Fset,
			conn:     conn,
			typeinfo: pkg.TypesInfo,
		}
		visitor.visit(file)
	}
}

func TraverseStmtAndExpr(conn *db.Connection, aresult Result) {
	pkg := aresult.Pkg
	for _, file := range pkg.Syntax {
		visitor := visitorExpr{
			func_ids: []int64{},
			file_set: pkg.Fset,
			conn:     conn,
			typeinfo: pkg.TypesInfo,
		}
		visitor.visit(file)
	}
}
