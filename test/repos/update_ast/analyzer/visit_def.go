package analyzer

import (
	"astdb/db"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
)

// visitor that updates "Function" and "Variable" tables
type visitorDef struct {
	func_ids []int64 // stack of func_id
	file_set *token.FileSet
	conn *db.Connection
	typeinfo *types.Info
}

func(v *visitorDef) name_pos(ident *ast.Ident) (string, token.Position){
	if ident==nil {return "", token.Position{}}
	return ident.Name, v.file_set.Position(ident.Pos())
}


func(vis *visitorDef) handle_gendecl(decl *ast.GenDecl, func_id int64, parent_id int64) {
	if decl.Tok != token.CONST && decl.Tok !=token.VAR {return}
	for _, spec := range decl.Specs {
		if v, ok := spec.(*ast.ValueSpec); ok {
			for _, name := range v.Names {
				_1, _2 := vis.name_pos(name)
				if _1 == "_" {continue}
				_, err := vis.conn.InsertIdentDef(_1, _2, func_id, parent_id, db.FieldOther, decl.Tok == token.CONST, vis.typeinfo.TypeOf(name))
				if err != nil {
					log.Printf("Fail to insert variable %v: %v", name, err)
				}
			}
		}
	}
}

func(vis *visitorDef) push(func_id int64) {
	vis.func_ids = append(vis.func_ids, func_id)
}

func(vis *visitorDef) pop() {
	vis.func_ids = vis.func_ids[0:len(vis.func_ids)-1]
}

func(vis *visitorDef) func_id() int64 {
	if len(vis.func_ids)==0 {return -1}
	return vis.func_ids[0]
}

func(vis *visitorDef) parent_id() int64 {
	if len(vis.func_ids)==0 {return -1}
	return vis.func_ids[len(vis.func_ids)-1]
}

func(vis *visitorDef) handle_func_type(functype *ast.FuncType, func_id int64, parent_id int64) {
	if functype.Params != nil {
		for _, p := range functype.Params.List {
			for _, name := range p.Names {
				_1, _2 := vis.name_pos(name)
				if _1 == "_" {continue}
				_, err := vis.conn.InsertIdentDef(_1, _2, func_id, parent_id, db.FieldParam, false, vis.typeinfo.TypeOf(name))
				if err != nil {
					log.Printf("Fail to insert variable %v: %v", name, err)
				}
			}
		}
	}
	if functype.Results != nil {
		for _, r := range functype.Results.List {
			for _, name := range r.Names {
				_1, _2 := vis.name_pos(name)
				if _1 == "_" {continue}
				_, err := vis.conn.InsertIdentDef(_1, _2, func_id, parent_id, db.FieldResult, false, vis.typeinfo.TypeOf(name))
				if err != nil {
					log.Printf("Fail to insert variable %v: %v", name, err)
				}
			}
		}
	}
}

func(vis *visitorDef) visit(node ast.Node) (_ unit) {
	if node==nil {return}
	switch node := node.(type) {
	// Comments and fields
	case *ast.Comment:
		// nothing to do

	case *ast.CommentGroup:
		for _, c := range node.List {
			vis.visit(c)
		}

	case *ast.Field:
		if node.Doc != nil {
			vis.visit(node.Doc)
		}
		walkNodeList[unit](vis, node.Names)
		if node.Type != nil {
			vis.visit(node.Type)
		}
		if node.Tag != nil {
			vis.visit(node.Tag)
		}
		if node.Comment != nil {
			vis.visit(node.Comment)
		}

	case *ast.FieldList:
		for _, f := range node.List {
			vis.visit(f)
		}

	// Expressions
	case *ast.BadExpr, *ast.Ident, *ast.BasicLit:
		// nothing to do

	case *ast.Ellipsis:
		if node.Elt != nil {
			vis.visit(node.Elt)
		}

	case *ast.FuncLit:
		sign := vis.typeinfo.TypeOf(node)
		if sign==nil {
			log.Printf("Fail to insert function literal at %s: sign is nil", vis.file_set.Position(node.Pos()))
			return
		}
		func_id, err := vis.conn.InsertFunction("", vis.file_set.Position(node.Pos()), sign.(*types.Signature),true,vis.parent_id())
		if err != nil {
			log.Printf("Fail to insert function literal at %s: %v", vis.file_set.Position(node.Pos()), err)
			return
		}
		vis.push(func_id)
		vis.handle_func_type(node.Type, vis.func_id(), vis.parent_id())
		if node.Body != nil {
			vis.visit(node.Body)
		}
		vis.pop()

	case *ast.CompositeLit:
		if node.Type != nil {
			vis.visit(node.Type)
		}
		walkNodeList[unit](vis, node.Elts)

	case *ast.ParenExpr:
		vis.visit(node.X)

	case *ast.SelectorExpr:
		vis.visit(node.X)
		vis.visit(node.Sel)

	case *ast.IndexExpr:
		vis.visit(node.X)
		vis.visit(node.Index)

	case *ast.IndexListExpr:
		vis.visit(node.X)
		for _, index := range node.Indices {
			vis.visit(index)
		}

	case *ast.SliceExpr:
		vis.visit(node.X)
		if node.Low != nil {
			vis.visit(node.Low)
		}
		if node.High != nil {
			vis.visit(node.High)
		}
		if node.Max != nil {
			vis.visit(node.Max)
		}

	case *ast.TypeAssertExpr:
		vis.visit(node.X)
		if node.Type != nil {
			vis.visit(node.Type)
		}

	case *ast.CallExpr:
		vis.visit(node.Fun)
		walkNodeList[unit](vis, node.Args)

	case *ast.StarExpr:
		vis.visit(node.X)

	case *ast.UnaryExpr:
		vis.visit(node.X)

	case *ast.BinaryExpr:
		vis.visit(node.X)
		vis.visit(node.Y)

	case *ast.KeyValueExpr:
		vis.visit(node.Key)
		vis.visit(node.Value)

	// Types
	case *ast.ArrayType:
		if node.Len != nil {
			vis.visit(node.Len)
		}
		vis.visit(node.Elt)

	case *ast.StructType:
		vis.visit(node.Fields)

	case *ast.FuncType:
		if node.TypeParams != nil {
			vis.visit(node.TypeParams)
		}
		if node.Params != nil {
			vis.visit(node.Params)
		}
		if node.Results != nil {
			vis.visit(node.Results)
		}

	case *ast.InterfaceType:
		vis.visit(node.Methods)

	case *ast.MapType:
		vis.visit(node.Key)
		vis.visit(node.Value)

	case *ast.ChanType:
		vis.visit(node.Value)

	// Statements
	case *ast.BadStmt:
		// nothing to do

	case *ast.DeclStmt:
		vis.visit(node.Decl)

	case *ast.EmptyStmt:
		// nothing to do

	case *ast.LabeledStmt:
		vis.visit(node.Label)
		vis.visit(node.Stmt)

	case *ast.ExprStmt:
		vis.visit(node.X)

	case *ast.SendStmt:
		vis.visit(node.Chan)
		vis.visit(node.Value)

	case *ast.IncDecStmt:
		vis.visit(node.X)

	case *ast.AssignStmt:
		if node.Tok == token.DEFINE {
			for _, lhs := range node.Lhs {
				if id, ok := lhs.(*ast.Ident); ok {
					if _, in := vis.typeinfo.Defs[id]; !in {continue}
					_1, _2 := vis.name_pos(id)
					if _1 == "_" {continue}
					_, err := vis.conn.InsertIdentDef(_1, _2, vis.func_id(), vis.parent_id(), db.FieldOther, false, vis.typeinfo.TypeOf(id))
					if err != nil {
						log.Printf("Fail to insert variable %v: %v", id, err)
					}
				}
			}
		}
		walkNodeList[unit](vis, node.Lhs)
		walkNodeList[unit](vis, node.Rhs)

	case *ast.GoStmt:
		vis.visit(node.Call)

	case *ast.DeferStmt:
		vis.visit(node.Call)

	case *ast.ReturnStmt:
		walkNodeList[unit](vis, node.Results)

	case *ast.BranchStmt:
		if node.Label != nil {
			vis.visit(node.Label)
		}

	case *ast.BlockStmt:
		walkNodeList[unit](vis, node.List)

	case *ast.IfStmt:
		if node.Init != nil {
			vis.visit(node.Init)
		}
		vis.visit(node.Cond)
		vis.visit(node.Body)
		if node.Else != nil {
			vis.visit(node.Else)
		}

	case *ast.CaseClause:
		walkNodeList[unit](vis, node.List)
		walkNodeList[unit](vis, node.Body)

	case *ast.SwitchStmt:
		if node.Init != nil {
			vis.visit(node.Init)
		}
		if node.Tag != nil {
			vis.visit(node.Tag)
		}
		vis.visit(node.Body)

	case *ast.TypeSwitchStmt:
		if node.Init != nil {
			vis.visit(node.Init)
		}
		vis.visit(node.Assign)
		vis.visit(node.Body)

	case *ast.CommClause:
		if node.Comm != nil {
			vis.visit(node.Comm)
		}
		walkNodeList[unit](vis, node.Body)

	case *ast.SelectStmt:
		vis.visit(node.Body)

	case *ast.ForStmt:
		if node.Init != nil {
			vis.visit(node.Init)
		}
		if node.Cond != nil {
			vis.visit(node.Cond)
		}
		if node.Post != nil {
			vis.visit(node.Post)
		}
		vis.visit(node.Body)

	case *ast.RangeStmt:
		if node.Key != nil {
			if id, ok := node.Key.(*ast.Ident); node.Tok==token.DEFINE && ok {
				if _, in := vis.typeinfo.Defs[id]; in {
					_1, _2 := vis.name_pos(id)
					if _1 != "_" {
						_, err := vis.conn.InsertIdentDef(_1, _2, vis.func_id(), vis.parent_id(), db.FieldOther, false, vis.typeinfo.TypeOf(id))
						if err != nil {
							log.Printf("Fail to insert variable %v: %v", id, err)
						}
					}
				}
			}
			vis.visit(node.Key)
		}
		if node.Value != nil {
			if id, ok := node.Value.(*ast.Ident); node.Tok==token.DEFINE &&  ok {
				if _, in := vis.typeinfo.Defs[id]; in {
					_1, _2 := vis.name_pos(id)
					if _1 != "_" {
						_, err := vis.conn.InsertIdentDef(_1, _2, vis.func_id(), vis.parent_id(), db.FieldOther, false, vis.typeinfo.TypeOf(id))
						if err != nil {
							log.Printf("Fail to insert variable %v: %v", id, err)
						}
					}
				}
			}
			vis.visit(node.Value)
		}
		vis.visit(node.X)
		vis.visit(node.Body)

	// Declarations
	case *ast.ImportSpec:
		if node.Name != nil {
			vis.visit(node.Name)
		}
		vis.visit(node.Path)

	case *ast.ValueSpec:
		walkNodeList[unit](vis, node.Names)
		if node.Type != nil {
			vis.visit(node.Type)
		}
		walkNodeList[unit](vis, node.Values)

	case *ast.TypeSpec:
		vis.visit(node.Name)
		if node.TypeParams != nil {
			vis.visit(node.TypeParams)
		}
		vis.visit(node.Type)

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		vis.handle_gendecl(node, vis.func_id(), vis.parent_id())
		for _, s := range node.Specs {
			vis.visit(s)
		}

	case *ast.FuncDecl:
		if len(vis.func_ids) != 0 {
			panic("nested function")
		}
		sign := vis.typeinfo.TypeOf(node.Name)
		if sign==nil {
			log.Printf("Fail to insert function %s: sign is nil", node.Name.Name)
			return
		}
		func_id, err := vis.conn.InsertFunction(node.Name.Name, vis.file_set.Position(node.Pos()), sign.(*types.Signature),false,-1)
		if err!=nil {
			log.Printf("Fail to insert function %s: %v", node.Name.Name, err)
			return
		}
		vis.push(func_id)

		vis.handle_func_type(node.Type, vis.func_id(), vis.parent_id())
		if node.Recv != nil {
			for _, r := range node.Recv.List {
				for _, name := range r.Names {
					_1, _2 := vis.name_pos(name)
					_, err := vis.conn.InsertIdentDef(_1, _2, vis.func_id(), vis.parent_id(), db.FieldReceiver, false, vis.typeinfo.TypeOf(name))
					if err != nil {
						log.Printf("Fail to insert variable %v: %v", name, err)
					}
				}
			}
		}

		if node.Body != nil {
			vis.visit(node.Body)
		}
		vis.pop()

	// Files and packages
	case *ast.File:
		walkNodeList[unit](vis, node.Decls)

	case *ast.Package:
		for _, f := range node.Files {
			vis.visit(f)
		}

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", node))
	}
	return
}

