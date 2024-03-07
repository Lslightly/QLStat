package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/tools/go/packages"
)

type memAccDepth struct {
	depth     int
	uncertain int
	ident     *ast.Ident
}

// depth: depth of indirect mem access. For example:
// ⟦Ident(a)⟧ = 0
// ⟦Dereference(a)⟧ = ⟦a⟧+1
// ⟦Reference(a)⟧ = ⟦a⟧-1
// ⟦Field-access(a,f)⟧ = ⟦a⟧
// ⟦Indirect-field-access(a,f)⟧ = ⟦a⟧+1
// ⟦Unknown-field-access(a,f)⟧ = ⟦a⟧+i
// ⟦Array-index(a,i)⟧ = ⟦a⟧
// ⟦Slice-index(a,i)⟧ = ⟦a⟧+1
// ⟦Unknown-index(a,f)⟧ = ⟦a⟧+i
// uncertain: Imaginary part
func getExprMemAccDepth(expr ast.Expr, pkg *packages.Package, cache *lru.Cache[token.Position, memAccDepth]) (depth memAccDepth) {
	if depth, ok := cache.Get(pkg.Fset.Position(expr.Pos())); ok {
		return depth
	}
	switch expr := expr.(type) {
	case *ast.Ident:
		depth = memAccDepth{0, 0, expr}
	case *ast.StarExpr:
		depth = getExprMemAccDepth(expr.X, pkg, cache)
		if depth.ident == nil {
			return
		}
		depth.depth++
	case *ast.UnaryExpr:
		depth = getExprMemAccDepth(expr.X, pkg, cache)
		if depth.ident == nil {
			return
		}
		if expr.Op == token.AND {
			depth.depth--
		} else if expr.Op == token.MUL {
			depth.depth++
		}
	case *ast.SelectorExpr:
		depth = getExprMemAccDepth(expr.X, pkg, cache)
		if depth.ident == nil {
			return
		}
		typ := pkg.TypesInfo.TypeOf(expr.X)
		if typ == nil {
			depth.uncertain++
		} else if _, ok := typ.Underlying().(*types.Pointer); ok {
			depth.depth++
		}
	case *ast.IndexExpr:
		depth = getExprMemAccDepth(expr.X, pkg, cache)
		if depth.ident == nil {
			return
		}
		typ := pkg.TypesInfo.TypeOf(expr.X)
		if typ == nil {
			depth.uncertain++
		} else if _, ok := typ.Underlying().(*types.Slice); ok {
			depth.depth++
		} else if _, ok := typ.Underlying().(*types.Map); ok {
			depth.depth++
		}
	case *ast.IndexListExpr:
		depth = getExprMemAccDepth(expr.X, pkg, cache)
		if depth.ident == nil {
			return
		}
		typ := pkg.TypesInfo.TypeOf(expr.X)
		if typ == nil {
			depth.uncertain++
		} else if _, ok := typ.Underlying().(*types.Slice); ok {
			depth.depth++
		}
	case *ast.ParenExpr:
		depth = getExprMemAccDepth(expr.X, pkg, cache)
		if depth.ident == nil {
			return
		}
	default:
		depth = memAccDepth{0, -1, nil}
		return
	}
	cache.Add(pkg.Fset.Position(expr.Pos()), depth)
	return
}

// inner expr:
// a: nil;
// *a, &a: a;
// a.f: a;
// a[i]: a;
// (a): a;
func getInnerExpr(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {
	case *ast.StarExpr:
		return expr.X
	case *ast.UnaryExpr:
		return expr.X
	case *ast.SelectorExpr:
		return expr.X
	case *ast.IndexExpr:
		return expr.X
	case *ast.IndexListExpr:
		return expr.X
	case *ast.ParenExpr:
		return expr.X
	default:
		return nil
	}
}
