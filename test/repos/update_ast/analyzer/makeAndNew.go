package analyzer

import (
	"astdb/db"
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"strconv"
)

/*
parse different format of `str` to int64, including hex, oct and decimal
*/
func parseManyFmtsInt(str string) (int64, error) {
	if str[0] == '0' {
		if len(str) == 1 {
			return 0, nil
		} else if str[1] == 'x' || str[1] == 'X' {
			return strconv.ParseInt(str[2:], 16, 64)
		} else {
			return strconv.ParseInt(str[1:], 8, 64)
		}
	} else {
		return strconv.ParseInt(str, 10, 64)
	}
}

/*
traverse all make() in `aresult.ASTPkg` and insert them to db `conn`
classify the makeslice, makemap, makechan
parse the capacity if possible, else use -1 to represent unknown capacity.
*/
func LightWeightTraverseMake(conn *db.Connection, aresult LightWeightResult) (err error) {
	pkg := aresult.ASTPkg
	files := pkg.Files
	var fileid int64
	fset := aresult.fset
	MakeFn := func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.File:
		case *ast.CallExpr:
			fnExpr := x.Fun
			switch fn := fnExpr.(type) {
			case *ast.Ident:
				if fn.Name == "make" {
					if len(x.Args) >= 1 {
						pos := fset.Position(x.Pos())
						typeExpr := x.Args[0]
						typeStr := ""
						// typeStr := fset.Position(typeExpr.Pos()).String()
						switch retType := typeExpr.(type) {
						case *ast.ArrayType:
							if retType.Len == nil {
								typeStr = "slice"
							}
						case *ast.MapType:
							typeStr = "map"
						case *ast.ChanType:
							typeStr = "chan"
						}
						if typeStr != "" {
							conn.InsertMake(db.Make{
								Fileid:  fileid,
								Line:    pos.Line,
								Column:  pos.Column,
								Typestr: typeStr,
								Cap:     getCap(x),
							})
						}
					}
					return false
				}
			}
		}
		return true
	}
	for path, astFile := range files {
		fileid, err = conn.FileID(path)
		if err != nil {
			fmt.Println(err)
			return err
		}
		ast.Inspect(astFile, MakeFn)
	}
	return nil
}

// the second or third argument(capacity) of make(T, ?)/make(T, ?, ?)
func getCap(x *ast.CallExpr) int64 {
	var capExpr ast.Expr = nil
	if len(x.Args) == 2 {
		capExpr = x.Args[1]
	} else if len(x.Args) == 3 {
		capExpr = x.Args[2]
	} else {
		return -1
	}
	if lit, ok := capExpr.(*ast.BasicLit); ok {
		if lit.Kind == token.INT {
			if cap, err := parseManyFmtsInt(lit.Value); err == nil {
				return cap
			}
		}
	}
	return -1
}

/*
traverse makeslice and new to get the size of object allocated
store the type string and the size
if size is unknown, -1 will be stored
*/
func TraverseMakesliceAndNew(conn *db.Connection, aresult Result) (err error) {
	pkg := aresult.Pkg
	var fileid int64
	fset := pkg.Fset
	astFiles := pkg.Syntax
	typeSizes := pkg.TypesSizes
	typeInfo := pkg.TypesInfo
	MakeAndNewFn := func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			pos := fset.Position(x.Pos())
			if fn, ok := x.Fun.(*ast.Ident); !ok {
				return true
			} else {
				switch fn.Name {
				case "make":
					if len(x.Args) < 1 {
						return false
					}
					typeExpr := x.Args[0]
					needInsert := false
					size := int64(1)
					switch retType := typeExpr.(type) {
					case *ast.ArrayType:
						if retType.Len == nil {
							elemTypeSize := typeSizes.Sizeof(typeInfo.TypeOf(retType.Elt))
							size *= elemTypeSize
							needInsert = true
						}
					}
					if needInsert {
						if cap := getCap(x); cap == -1 {
							size = -1
						} else {
							size *= cap
						}
						conn.InsertMakeslice(db.Makeslice{
							Fileid:  fileid,
							Line:    pos.Line,
							Column:  pos.Column,
							TypeStr: printExpr(typeExpr, fset),
							Size:    size,
						})
					}
					return false
				case "new":
					typeExpr := x.Args[0]
					size := typeSizes.Sizeof(typeInfo.TypeOf(typeExpr))
					conn.InsertNew(db.New{
						Fileid:  fileid,
						Line:    pos.Line,
						Column:  pos.Column,
						TypeStr: printExpr(typeExpr, fset),
						Size:    size,
					})
					return false
				}
			}
		}
		return true
	}
	for _, astFile := range astFiles {
		fileName := fset.Position(astFile.Pos()).Filename
		fileid, err = conn.FileID(fileName)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer func() {
			if v := recover(); v != nil {
				fmt.Println(v)
			}
		}()
		ast.Inspect(astFile, MakeAndNewFn)
	}
	return
}

func printExpr(expr ast.Expr, fset *token.FileSet) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, fset, expr)
	return buf.String()
}
