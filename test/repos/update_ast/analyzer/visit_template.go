package analyzer

import (
	"fmt"
	"go/ast"
)

// a template for visitor
// we choose not to use more sophisticated appoaches like visitor pattern
// for simplicity of coding
type visitorTemplate struct {

}

func(v *visitorTemplate) visit(node ast.Node) (_ unit) {
	switch n := node.(type) {
	// Comments and fields
	case *ast.Comment:
		// nothing to do

	case *ast.CommentGroup:
		for _, c := range n.List {
			v.visit(c)
		}

	case *ast.Field:
		if n.Doc != nil {
			v.visit(n.Doc)
		}
		walkNodeList[unit](v, n.Names)
		if n.Type != nil {
			v.visit(n.Type)
		}
		if n.Tag != nil {
			v.visit(n.Tag)
		}
		if n.Comment != nil {
			v.visit(n.Comment)
		}

	case *ast.FieldList:
		for _, f := range n.List {
			v.visit(f)
		}

	// Expressions
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	case *ast.Ellipsis:
		if n.Elt != nil {
			v.visit(n.Elt)
		}

	case *ast.FuncLit:
		v.visit(n.Type)
		v.visit(n.Body)

	case *ast.CompositeLit:
		if n.Type != nil {
			v.visit(n.Type)
		}
		walkNodeList[unit](v, n.Elts)

	case *ast.ParenExpr:
		v.visit(n.X)

	case *ast.SelectorExpr:
		v.visit(n.X)
		v.visit(n.Sel)

	case *ast.IndexExpr:
		v.visit(n.X)
		v.visit(n.Index)

	case *ast.IndexListExpr:
		v.visit(n.X)
		for _, index := range n.Indices {
			v.visit(index)
		}

	case *ast.SliceExpr:
		v.visit(n.X)
		if n.Low != nil {
			v.visit(n.Low)
		}
		if n.High != nil {
			v.visit(n.High)
		}
		if n.Max != nil {
			v.visit(n.Max)
		}

	case *ast.TypeAssertExpr:
		v.visit(n.X)
		if n.Type != nil {
			v.visit(n.Type)
		}

	case *ast.CallExpr:
		v.visit(n.Fun)
		walkNodeList[unit](v, n.Args)

	case *ast.StarExpr:
		v.visit(n.X)

	case *ast.UnaryExpr:
		v.visit(n.X)

	case *ast.BinaryExpr:
		v.visit(n.X)
		v.visit(n.Y)

	case *ast.KeyValueExpr:
		v.visit(n.Key)
		v.visit(n.Value)

	// Types
	case *ast.ArrayType:
		if n.Len != nil {
			v.visit(n.Len)
		}
		v.visit(n.Elt)

	case *ast.StructType:
		v.visit(n.Fields)

	case *ast.FuncType:
		if n.TypeParams != nil {
			v.visit(n.TypeParams)
		}
		if n.Params != nil {
			v.visit(n.Params)
		}
		if n.Results != nil {
			v.visit(n.Results)
		}

	case *ast.InterfaceType:
		v.visit(n.Methods)

	case *ast.MapType:
		v.visit(n.Key)
		v.visit(n.Value)

	case *ast.ChanType:
		v.visit(n.Value)

	// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		v.visit(n.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		v.visit(n.Label)
		v.visit(n.Stmt)

	case *ast.ExprStmt:
		v.visit(n.X)

	case *ast.SendStmt:
		v.visit(n.Chan)
		v.visit(n.Value)

	case *ast.IncDecStmt:
		v.visit(n.X)

	case *ast.AssignStmt:
		walkNodeList[unit](v, n.Lhs)
		walkNodeList[unit](v, n.Rhs)

	case *ast.GoStmt:
		v.visit(n.Call)

	case *ast.DeferStmt:
		v.visit(n.Call)

	case *ast.ReturnStmt:
		walkNodeList[unit](v, n.Results)

	case *ast.BranchStmt:
		if n.Label != nil {
			v.visit(n.Label)
		}

	case *ast.BlockStmt:
		walkNodeList[unit](v, n.List)

	case *ast.IfStmt:
		if n.Init != nil {
			v.visit(n.Init)
		}
		v.visit(n.Cond)
		v.visit(n.Body)
		if n.Else != nil {
			v.visit(n.Else)
		}

	case *ast.CaseClause:
		walkNodeList[unit](v, n.List)
		walkNodeList[unit](v, n.Body)

	case *ast.SwitchStmt:
		if n.Init != nil {
			v.visit(n.Init)
		}
		if n.Tag != nil {
			v.visit(n.Tag)
		}
		v.visit(n.Body)

	case *ast.TypeSwitchStmt:
		if n.Init != nil {
			v.visit(n.Init)
		}
		v.visit(n.Assign)
		v.visit(n.Body)

	case *ast.CommClause:
		if n.Comm != nil {
			v.visit(n.Comm)
		}
		walkNodeList[unit](v, n.Body)

	case *ast.SelectStmt:
		v.visit(n.Body)

	case *ast.ForStmt:
		if n.Init != nil {
			v.visit(n.Init)
		}
		if n.Cond != nil {
			v.visit(n.Cond)
		}
		if n.Post != nil {
			v.visit(n.Post)
		}
		v.visit(n.Body)

	case *ast.RangeStmt:
		if n.Key != nil {
			v.visit(n.Key)
		}
		if n.Value != nil {
			v.visit(n.Value)
		}
		v.visit(n.X)
		v.visit(n.Body)

	// Declarations
	case *ast.ImportSpec:
		if n.Doc != nil {
			v.visit(n.Doc)
		}
		if n.Name != nil {
			v.visit(n.Name)
		}
		v.visit(n.Path)
		if n.Comment != nil {
			v.visit(n.Comment)
		}

	case *ast.ValueSpec:
		if n.Doc != nil {
			v.visit(n.Doc)
		}
		walkNodeList[unit](v, n.Names)
		if n.Type != nil {
			v.visit(n.Type)
		}
		walkNodeList[unit](v, n.Values)
		if n.Comment != nil {
			v.visit(n.Comment)
		}

	case *ast.TypeSpec:
		if n.Doc != nil {
			v.visit(n.Doc)
		}
		v.visit(n.Name)
		if n.TypeParams != nil {
			v.visit(n.TypeParams)
		}
		v.visit(n.Type)
		if n.Comment != nil {
			v.visit(n.Comment)
		}

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		if n.Doc != nil {
			v.visit(n.Doc)
		}
		for _, s := range n.Specs {
			v.visit(s)
		}

	case *ast.FuncDecl:
		if n.Doc != nil {
			v.visit(n.Doc)
		}
		if n.Recv != nil {
			v.visit(n.Recv)
		}
		v.visit(n.Name)
		v.visit(n.Type)
		if n.Body != nil {
			v.visit(n.Body)
		}

	// Files and packages
	case *ast.File:
		if n.Doc != nil {
			v.visit(n.Doc)
		}
		v.visit(n.Name)
		walkNodeList[unit](v, n.Decls)
		// don't walk n.Comments - they have been
		// visited already through the individual
		// nodes

	case *ast.Package:
		for _, f := range n.Files {
			v.visit(f)
		}

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}
	return
}
