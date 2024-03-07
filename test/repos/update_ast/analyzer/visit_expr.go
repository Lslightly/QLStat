package analyzer

import (
	"astdb/db"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"
)

// visitor that updates "Expression" and "Statement" tables
type visitorExpr struct {
	func_ids       []int64 // stack of func_id
	file_set       *token.FileSet
	conn           *db.Connection
	typeinfo       *types.Info
	not_mem_access bool // for internal use
}

type exprResult struct {
	id                int64
	depth             int
	depth_uncertainty int
	base_var_name     string
	base_var_id       int64
	is_mem_acc        bool
}

func (vis *visitorExpr) push(func_id int64) {
	vis.func_ids = append(vis.func_ids, func_id)
}

func (vis *visitorExpr) pop() {
	vis.func_ids = vis.func_ids[0 : len(vis.func_ids)-1]
}

func (vis *visitorExpr) func_id() int64 {
	if len(vis.func_ids) == 0 {
		return -1
	}
	return vis.func_ids[0]
}

func (vis *visitorExpr) parent_id() int64 {
	if len(vis.func_ids) == 0 {
		return -1
	}
	return vis.func_ids[len(vis.func_ids)-1]
}

func (v *visitorExpr) visit(node ast.Node) exprResult {
	not_mem_access := v.not_mem_access
	v.not_mem_access = false
	idOnly := func(id int64) exprResult {
		return exprResult{id: id}
	}
	if node == nil {
		return idOnly(-1)
	}
	switch n := node.(type) {
	// Comments and fields
	case *ast.Comment:
		// nothing to do

	case *ast.CommentGroup:
		// nothing to do

	case *ast.Field:
		walkNodeList[exprResult](v, n.Names)
		if n.Type != nil {
			v.visit(n.Type)
		}

	case *ast.FieldList:
		for _, f := range n.List {
			v.visit(f)
		}

	// Expressions
	case *ast.BadExpr:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.BadExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		return idOnly(expr_id)

	case *ast.BasicLit:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.BasicLitExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		return idOnly(expr_id)

	case *ast.Ident:
		expr_id, inserted, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.IdentExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		obj := v.typeinfo.ObjectOf(n)
		if inserted {
			if obj != nil {
				pos := v.file_set.Position(obj.Pos())
				err = v.conn.InsertExprIdent(expr_id, n.Name, &pos)
			} else {
				err = v.conn.InsertExprIdent(expr_id, n.Name, nil)
			}
			if err != nil {
				log.Panicf("Error: fail to insert expression %v: %v", n, err)
				return idOnly(-1)
			}
		}
		// set by ast.SelectorExpr, indicating that n is SelectorExpr.Sel
		if not_mem_access {
			return idOnly(expr_id)
		}

		if obj == nil {
			return idOnly(expr_id)
		}
		// skip pkgName, const, typename, etc.
		if variable, ok := obj.(*types.Var); !ok {
			return idOnly(expr_id)
			// skip field
		} else if variable.IsField() {
			return idOnly(expr_id)
		}
		// skip define
		if obj.Pos() == n.Pos() {
			return idOnly(expr_id)
		}
		base_var_id, err := v.conn.QueryIdentDef(v.file_set.Position(obj.Pos()))
		if err != nil {
			log.Printf("Error: fail to find identifier definition (0): name=%s,pos=%v: %v", obj.Name(), v.file_set.Position(obj.Pos()), err)
			base_var_id = -1
		}
		if inserted {
			err = v.conn.InsertExprMemAcc(expr_id, db.Ident, -1, n.Name, base_var_id, 1, 0)
			if err != nil {
				log.Panicf("Error: fail to insert mem access %v: %v", n, err)
			}
		}
		return exprResult{
			id:                expr_id,
			depth:             1,
			depth_uncertainty: 0,
			base_var_name:     n.Name,
			base_var_id:       base_var_id,
			is_mem_acc:        true,
		}

	case *ast.Ellipsis:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.EllipsisExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Elt)
		return idOnly(expr_id)

	case *ast.FuncLit:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.FuncLitExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		func_id, err := v.conn.QueryFunction(v.file_set.Position(n.Pos()))
		if err != nil {
			log.Printf("Fail to find function literal at %s: %v", v.file_set.Position(n.Pos()), err)
			return idOnly(expr_id)
		}
		v.push(func_id)
		v.visit(n.Body)
		v.pop()
		return idOnly(expr_id)

	case *ast.CompositeLit:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.CompositeLitExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		if n.Type != nil {
			v.visit(n.Type)
		}
		walkNodeList[exprResult](v, n.Elts)
		return idOnly(expr_id)

	case *ast.ParenExpr:
		expr_id, inserted, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.ParenExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		result := v.visit(n.X)
		if inserted && result.is_mem_acc {
			err = v.conn.InsertExprMemAcc(expr_id, db.Paren, result.id, result.base_var_name, result.base_var_id, result.depth, result.depth_uncertainty)
			if err != nil {
				panic(err)
			}
		}
		result.id = expr_id
		return result

	case *ast.SelectorExpr:
		expr_id, inserted, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.SlectorExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		result := v.visit(n.X)
		v.not_mem_access = true
		v.visit(n.Sel)
		if obj_sel := v.typeinfo.ObjectOf(n.Sel); obj_sel != nil {
			if obj_sel_as_var, ok := obj_sel.(*types.Var); ok {
				if result.is_mem_acc {
					typ := v.typeinfo.TypeOf(n.X)
					delta_depth := 0
					delta_depth_uncertain := 0
					if typ == nil {
						log.Printf("Warning: unknown field access at %v", v.file_set.Position(n.Pos()))
						delta_depth_uncertain++
						if inserted {
							err = v.conn.InsertExprMemAcc(expr_id, db.UnknownFieldAccess, result.id, result.base_var_name, result.base_var_id, result.depth, result.depth_uncertainty+1)
						}
					} else if _, ok := typ.Underlying().(*types.Pointer); ok {
						delta_depth++
						if inserted {
							err = v.conn.InsertExprMemAcc(expr_id, db.IndirectFiledAccess, result.id, result.base_var_name, result.base_var_id, result.depth+1, result.depth_uncertainty)
						}
					} else if _, ok := typ.Underlying().(*types.Struct); ok {
						if inserted {
							err = v.conn.InsertExprMemAcc(expr_id, db.FieldAccess, result.id, result.base_var_name, result.base_var_id, result.depth, result.depth_uncertainty)
						}
					} else {
						return idOnly(expr_id)
					}
					if err != nil {
						log.Panicf("Error: fail to insert mem access %v: %v", n, err)
					}
					return exprResult{
						id:                expr_id,
						depth:             result.depth + delta_depth,
						depth_uncertainty: result.depth_uncertainty + delta_depth_uncertain,
						base_var_name:     result.base_var_name,
						base_var_id:       result.base_var_id,
						is_mem_acc:        true,
					}
				} else if ident_x, ok := n.X.(*ast.Ident); ok && !obj_sel_as_var.IsField() {
					if obj_x := v.typeinfo.ObjectOf(ident_x); obj_x != nil {
						if _, ok := obj_x.(*types.PkgName); ok {
							pos := v.file_set.Position(obj_sel.Pos())
							var base_var_id int64 = -1
							// only attempt to find an identifier from the same module
							if strings.HasPrefix(pos.Filename, RepoDir+"/") {
								base_var_id, err = v.conn.QueryIdentDef(pos)
								if err != nil {
									log.Printf("Error: fail to find identifier definition (1): name=%s,pos=%v: %v", obj_sel, pos, err)
									base_var_id = -1
								}
							}
							if inserted {
								err = v.conn.InsertExprMemAcc(expr_id, db.Ident, -1, n.Sel.Name, base_var_id, 1, 0)
								if err != nil {
									log.Panicf("Error: fail to insert mem access %v: %v", n, err)
								}
							}
							return exprResult{
								id:                expr_id,
								depth:             1,
								depth_uncertainty: 0,
								base_var_name:     n.Sel.Name,
								base_var_id:       base_var_id,
								is_mem_acc:        true,
							}
						}
					}
				}
			}
		}
		return idOnly(expr_id)

	case *ast.IndexExpr:
		expr_id, inserted, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.IndexExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		result := v.visit(n.X)
		v.visit(n.Index)
		if result.is_mem_acc {
			typ := v.typeinfo.TypeOf(n.X)
			delta_depth := 0
			delta_depth_uncertain := 0
			if typ == nil {
				log.Printf("Warning: unknown field access at %v", v.file_set.Position(n.Pos()))
				delta_depth_uncertain++
				if inserted {
					err = v.conn.InsertExprMemAcc(expr_id, db.UnknownIndex, result.id, result.base_var_name, result.base_var_id, result.depth, result.depth_uncertainty+1)
				}
			} else if _, ok := typ.Underlying().(*types.Slice); ok {
				delta_depth++
				if inserted {
					err = v.conn.InsertExprMemAcc(expr_id, db.SliceIndex, result.id, result.base_var_name, result.base_var_id, result.depth+1, result.depth_uncertainty)
				}
			} else if _, ok := typ.Underlying().(*types.Array); ok {
				if inserted {
					err = v.conn.InsertExprMemAcc(expr_id, db.ArrayIndex, result.id, result.base_var_name, result.base_var_id, result.depth, result.depth_uncertainty)
				}
			} else if _, ok := typ.Underlying().(*types.Map); ok {
				delta_depth++
				if inserted {
					err = v.conn.InsertExprMemAcc(expr_id, db.MapIndex, result.id, result.base_var_name, result.base_var_id, result.depth, result.depth_uncertainty)
				}
			} else {
				return idOnly(expr_id)
			}
			if err != nil {
				log.Panicf("Error: fail to insert mem access %v: %v", n, err)
			}
			return exprResult{
				id:                expr_id,
				depth:             result.depth + delta_depth,
				depth_uncertainty: result.depth_uncertainty + delta_depth_uncertain,
				base_var_name:     result.base_var_name,
				base_var_id:       result.base_var_id,
				is_mem_acc:        true,
			}
		}
		return idOnly(expr_id)

	case *ast.IndexListExpr:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.IndexListExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.X)
		for _, index := range n.Indices {
			v.visit(index)
		}
		return idOnly(expr_id)

	case *ast.SliceExpr:
		expr_id, inserted, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.SliceExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		result := v.visit(n.X)
		if n.Low != nil {
			v.visit(n.Low)
		}
		if n.High != nil {
			v.visit(n.High)
		}
		if n.Max != nil {
			v.visit(n.Max)
		}
		if result.is_mem_acc {
			typ := v.typeinfo.TypeOf(n.X)
			delta_depth := 0
			if typ == nil {
				log.Printf("Warning: unknown field access at %v", v.file_set.Position(n.Pos()))
				if inserted {
					err = v.conn.InsertExprMemAcc(expr_id, db.UnknownSlice, result.id, result.base_var_name, result.base_var_id, result.depth, result.depth_uncertainty+1)
				}
			} else if _, ok := typ.Underlying().(*types.Slice); ok {
				if inserted {
					err = v.conn.InsertExprMemAcc(expr_id, db.SliceSlice, result.id, result.base_var_name, result.base_var_id, result.depth, result.depth_uncertainty)
				}
			} else if _, ok := typ.Underlying().(*types.Array); ok {
				delta_depth--
				if inserted {
					err = v.conn.InsertExprMemAcc(expr_id, db.ArraySlice, result.id, result.base_var_name, result.base_var_id, result.depth-1, result.depth_uncertainty)
				}
			} else {
				return idOnly(expr_id)
			}
			if err != nil {
				log.Panicf("Error: fail to insert mem access %v: %v", n, err)
			}
			return exprResult{
				id:                expr_id,
				depth:             result.depth + delta_depth,
				depth_uncertainty: result.depth_uncertainty,
				base_var_name:     result.base_var_name,
				base_var_id:       result.base_var_id,
				is_mem_acc:        true,
			}
		}
		return idOnly(expr_id)

	case *ast.TypeAssertExpr:
		expr_id, inserted, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.TypeAssertExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		base_id := v.visit(n.X)
		var type_id int64 = -1
		if n.Type != nil {
			v.visit(n.Type)
			if inserted {
				var err error
				type_id, err = v.conn.TypeIDCheckNil(v.typeinfo.TypeOf(n.Type))
				if err != nil {
					log.Printf("Error: fail to find type id %v: %v", v.file_set.Position(n.Pos()), err)
					type_id = -1
				}
			}
		}
		if inserted {
			err = v.conn.InsertExprTypeAssert(expr_id, base_id.id, type_id)
			if err != nil {
				log.Panicf("Error: fail to insert type assert expression %v: %v", n, err)
			}
		}
		return idOnly(expr_id)

	case *ast.CallExpr:
		expr_id, inserted, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.CallExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		func_result := v.visit(n.Fun)
		arg_ids := walkNodeList[exprResult](v, n.Args)
		if inserted {
			for pos, arg := range arg_ids {
				err := v.conn.InsertExprCall(expr_id, func_result.id, arg.id, pos)
				if err != nil {
					log.Panicf("Error: fail to insert func call expression %v: %v", n, err)
				}
			}
			err = v.conn.InsertExprMemAcc(expr_id, db.FuncCall, -1, "", -1, 0, 0)
			if err != nil {
				log.Panicf("Error: fail to insert mem access %v: %v", n, err)
			}
		}
		return exprResult{
			id:                expr_id,
			depth:             0,
			depth_uncertainty: 0,
			base_var_name:     "",
			base_var_id:       -1,
			is_mem_acc:        true,
		}

	case *ast.StarExpr:
		expr_id, inserted, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.StarExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		result := v.visit(n.X)
		if result.is_mem_acc {
			if inserted {
				err = v.conn.InsertExprMemAcc(expr_id, db.Deref, result.id, result.base_var_name, result.base_var_id, result.depth+1, result.depth_uncertainty)
				if err != nil {
					log.Panicf("Error: fail to insert mem access %v: %v", n, err)
				}
			}
			result.id = expr_id
			result.depth++
			return result
		}
		return idOnly(expr_id)

	case *ast.UnaryExpr:
		expr_id, inserted, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.UnaryExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		result := v.visit(n.X)
		if result.is_mem_acc && n.Op == token.AND {
			if inserted {
				err = v.conn.InsertExprMemAcc(expr_id, db.Ref, result.id, result.base_var_name, result.base_var_id, result.depth-1, result.depth_uncertainty)
				if err != nil {
					log.Panicf("Error: fail to insert mem access %v: %v", n, err)
				}
			}
			result.id = expr_id
			result.depth--
			return result
		}
		return idOnly(expr_id)

	case *ast.BinaryExpr:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.BinaryExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.X)
		v.visit(n.Y)
		return idOnly(expr_id)

	case *ast.KeyValueExpr:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.KeyValueExpr)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Key)
		v.visit(n.Value)
		return idOnly(expr_id)

	// Types
	case *ast.ArrayType:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.ArrayType)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		if n.Len != nil {
			v.visit(n.Len)
		}
		v.visit(n.Elt)
		return idOnly(expr_id)

	case *ast.StructType:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.StructType)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Fields)
		return idOnly(expr_id)

	case *ast.FuncType:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.FuncType)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		if n.TypeParams != nil {
			v.visit(n.TypeParams)
		}
		if n.Params != nil {
			v.visit(n.Params)
		}
		if n.Results != nil {
			v.visit(n.Results)
		}
		return idOnly(expr_id)

	case *ast.InterfaceType:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.InterfaceType)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Methods)
		return idOnly(expr_id)

	case *ast.MapType:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.MapType)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Key)
		v.visit(n.Value)
		return idOnly(expr_id)

	case *ast.ChanType:
		expr_id, _, err := v.conn.InsertExpr(v.func_id(), v.parent_id(), v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.typeinfo.TypeOf(n), db.ChanType)
		if err != nil {
			log.Printf("Error: fail to insert expression %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Value)
		return idOnly(expr_id)

	// Statements
	case *ast.BadStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.BadStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		return idOnly(stmt_id)

	case *ast.DeclStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.DeclStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Decl)
		return idOnly(stmt_id)

	case *ast.EmptyStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.EmptyStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		return idOnly(stmt_id)

	case *ast.LabeledStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.LabeledStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Label)
		v.visit(n.Stmt)
		return idOnly(stmt_id)

	case *ast.ExprStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.ExprStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.X)
		return idOnly(stmt_id)

	case *ast.SendStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.SendStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Chan)
		v.visit(n.Value)
		return idOnly(stmt_id)

	case *ast.IncDecStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.IncDecStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.X)
		return idOnly(stmt_id)

	case *ast.AssignStmt:
		stmt_id, inserted, err := v.conn.InsertStmt(db.AssignStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		lhs_result := walkNodeList[exprResult](v, n.Lhs)
		rhs_result := walkNodeList[exprResult](v, n.Rhs)
		if inserted {
			for pos, lhs := range lhs_result {
				err = v.conn.InsertStmtLhs(stmt_id, pos, lhs.id)
				if err != nil {
					panic(err)
				}
			}
			for pos, rhs := range rhs_result {
				err = v.conn.InsertStmtRhs(stmt_id, pos, rhs.id)
				if err != nil {
					panic(err)
				}
			}
		}
		return idOnly(stmt_id)

	case *ast.GoStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.GoStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Call)
		return idOnly(stmt_id)

	case *ast.DeferStmt:
		stmt_id, inserted, err := v.conn.InsertStmt(db.DeferStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		expr_id := v.visit(n.Call).id
		if inserted {
			err = v.conn.InsertStmtDefer(stmt_id, expr_id)
			if err != nil {
				panic(err)
			}
		}
		return idOnly(stmt_id)

	case *ast.ReturnStmt:
		stmt_id, inserted, err := v.conn.InsertStmt(db.ReturnStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		ret_results := walkNodeList[exprResult](v, n.Results)
		if inserted {
			for pos, ret := range ret_results {
				err = v.conn.InsertStmtRetRes(stmt_id, pos, ret.id)
				if err != nil {
					panic(err)
				}
			}
		}
		return idOnly(stmt_id)

	case *ast.BranchStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.BranchStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		if n.Label != nil {
			v.visit(n.Label)
		}
		return idOnly(stmt_id)

	case *ast.BlockStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.BlockStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		walkNodeList[exprResult](v, n.List)
		return idOnly(stmt_id)

	case *ast.IfStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.IfStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		if n.Init != nil {
			v.visit(n.Init)
		}
		v.visit(n.Cond)
		v.visit(n.Body)
		if n.Else != nil {
			v.visit(n.Else)
		}
		return idOnly(stmt_id)

	case *ast.CaseClause:
		stmt_id, _, err := v.conn.InsertStmt(db.CaseClause, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		walkNodeList[exprResult](v, n.List)
		walkNodeList[exprResult](v, n.Body)
		return idOnly(stmt_id)

	case *ast.SwitchStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.SwitchStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		if n.Init != nil {
			v.visit(n.Init)
		}
		if n.Tag != nil {
			v.visit(n.Tag)
		}
		v.visit(n.Body)
		return idOnly(stmt_id)

	case *ast.TypeSwitchStmt:
		stmt_id, inserted, err := v.conn.InsertStmt(db.TypeSwitchStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		if n.Init != nil {
			v.visit(n.Init)
		}
		v.visit(n.Assign)
		if inserted {
			to_name := ""
			var assert_expr *ast.TypeAssertExpr
			if assign, ok := n.Assign.(*ast.AssignStmt); ok {
				to_name = assign.Lhs[0].(*ast.Ident).Name
				assert_expr = assign.Rhs[0].(*ast.TypeAssertExpr)
			} else {
				assert_expr = n.Assign.(*ast.ExprStmt).X.(*ast.TypeAssertExpr)
			}
			assert_id, err := v.conn.QueryExpr(v.file_set.Position(assert_expr.Pos()), v.file_set.Position(assert_expr.End()))
			if err != nil {
				panic(err)
			}
			err = v.conn.InsertTypeSwitch(stmt_id, to_name, assert_id)
			if err != nil {
				panic(err)
			}
		}
		v.visit(n.Body)
		if inserted {
			for pos0, stmt0 := range n.Body.List {
				stmt := stmt0.(*ast.CaseClause)
				for pos1, typ := range stmt.List {
					type_id, err := v.conn.TypeID(v.typeinfo.TypeOf(typ))
					if err != nil {
						log.Printf("Error: fail to find type id %v: %v", v.file_set.Position(n.Pos()), err)
						type_id = -1
					}
					err = v.conn.InsertTypeSwitchCase(stmt_id, pos0, pos1, type_id)
					if err != nil {
						panic(err)
					}
				}
			}
		}
		return idOnly(stmt_id)

	case *ast.CommClause:
		stmt_id, _, err := v.conn.InsertStmt(db.CommClause, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		if n.Comm != nil {
			v.visit(n.Comm)
		}
		walkNodeList[exprResult](v, n.Body)
		return idOnly(stmt_id)

	case *ast.SelectStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.SelectStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		v.visit(n.Body)
		return idOnly(stmt_id)

	case *ast.ForStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.ForStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
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
		return idOnly(stmt_id)

	case *ast.RangeStmt:
		stmt_id, _, err := v.conn.InsertStmt(db.RangeStmt, v.file_set.Position(n.Pos()), v.file_set.Position(n.End()), v.func_id(), v.parent_id())
		if err != nil {
			log.Printf("Error: fail to insert statement %v: %v", n, err)
			return idOnly(-1)
		}
		if n.Key != nil {
			v.visit(n.Key)
		}
		if n.Value != nil {
			v.visit(n.Value)
		}
		v.visit(n.X)
		v.visit(n.Body)
		return idOnly(stmt_id)

	// Declarations
	case *ast.ImportSpec:
		if n.Name != nil {
			v.visit(n.Name)
		}
		v.visit(n.Path)

	case *ast.ValueSpec:
		walkNodeList[exprResult](v, n.Names)
		if n.Type != nil {
			v.visit(n.Type)
		}
		walkNodeList[exprResult](v, n.Values)

	case *ast.TypeSpec:
		v.visit(n.Name)
		if n.TypeParams != nil {
			v.visit(n.TypeParams)
		}
		v.visit(n.Type)

	case *ast.BadDecl:
		// nothing to do

	case *ast.GenDecl:
		for _, s := range n.Specs {
			v.visit(s)
		}

	case *ast.FuncDecl:
		func_id, err := v.conn.QueryFunction(v.file_set.Position(n.Pos()))
		if err != nil {
			log.Printf("Fail to find function at %s: %v", v.file_set.Position(n.Pos()), err)
			return idOnly(-1)
		}
		v.push(func_id)
		if n.Recv != nil {
			v.visit(n.Recv)
		}
		v.visit(n.Name)
		v.visit(n.Type)
		if n.Body != nil {
			v.visit(n.Body)
		}
		v.pop()

	// Files and packages
	case *ast.File:
		v.visit(n.Name)
		walkNodeList[exprResult](v, n.Decls)

	case *ast.Package:
		for _, f := range n.Files {
			v.visit(f)
		}

	default:
		panic(fmt.Sprintf("ast.Walk: unexpected node type %T", n))
	}
	return idOnly(-1)
}
