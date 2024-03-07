package analyzer

import (
	"go/ast"
	"go/token"
	"log"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/tools/go/packages"
)

const CACHE_SIZE = 65536

func newCache[Key comparable, Value any](size int) *lru.Cache[Key,Value] {
	cache, err := lru.New[Key,Value](size)
	if err != nil {
		log.Fatal(err)
	}
	return cache
}

func namePos(ident *ast.Ident, pkg *packages.Package) (string, token.Position){
	return ident.Name, pkg.Fset.Position(ident.Pos())
}

func nameDefPos(ident *ast.Ident, pkg *packages.Package) token.Position {
	obj := ident.Obj
	if obj == nil || obj.Decl == nil {
		return pkg.Fset.Position(ident.Pos())
	}
	switch decl := obj.Decl.(type) {
		case *ast.Field:
			for _, n := range decl.Names {
				if n.Name == obj.Name {
					return pkg.Fset.Position(n.Pos())
				}
			}
		case *ast.ValueSpec:
			for _, n := range decl.Names {
				if n.Name == obj.Name {
					return pkg.Fset.Position(n.Pos())
				}
			}
		case *ast.TypeSpec:
			if decl.Name.Name == obj.Name {
				return pkg.Fset.Position(decl.Name.Pos())
			}
		case *ast.AssignStmt:
			for _, x := range decl.Lhs {
				if ident, isIdent := x.(*ast.Ident); isIdent && ident.Name == obj.Name {
					return pkg.Fset.Position(ident.Pos())
				}
			}
		case *ast.FuncDecl:
			if decl.Name.Name == obj.Name {
				return pkg.Fset.Position(decl.Name.Pos())
			}			
	}
	return pkg.Fset.Position(ident.Pos())
}