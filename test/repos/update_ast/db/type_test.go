package db_test

import (
	"fmt"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestType(t *testing.T) {
	path := "/data/github_go/repos/zap"
	cfg := packages.Config {
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports |
		packages.NeedDeps| packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax,
		Dir: path,
	}
	pkgs, err := packages.Load(&cfg, path)
	if err != nil {
		t.Fatal(err)
	}
	for _,pkg := range pkgs {
		t.Log("PACKAGE: ", pkg.Name)
		types := make(map[string]bool)
		for _, tp := range pkg.TypesInfo.Types {
			if tp.IsType() {
				types[fmt.Sprintf("%T: %v", tp.Type, tp.Type)] = true
			}
		}
		for tp := range types {
			t.Log(tp)
		}
	}
}
