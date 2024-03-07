package db

import (
	"go/ast"
	"go/token"
	"go/types"
)

type MemAccKind int

// MemAccKind: mem access mode (Nullable)
// 0: Ident: A （A is an identifier）
// 1: Deref: *A, and A is mem access
// 2: Ref: &A, and A is mem access
// 3: FieldAccess: A.f, A is mem access and the underlying type of A is struct type。
// 4: IndirectFiledAccess: A.f, a is mem access and A is pointer
// 5: SliceIndex: A[i], A is mem access, and type of A is slice type
// 6: ArrayIndex: A[i], A is mem access, and type of A is array type
// 7: MapIndex: A[i], A is mem access, and type of A is map type
// 8: Paren: (A)
// 9: UnknownFieldAccess A.f, but type of A is unknown
// 10: UnknownIndex: A[i], but type of A is unknown
// 11: OtherMemAcc: other mem access
// 12: Array slice: a[i:j]
// 13: Unknown slice
// 14: FuncCall: foo(???)
const (
	OtherMemAcc MemAccKind = iota - 1
	Ident
	Deref
	Ref
	FieldAccess
	IndirectFiledAccess
	SliceIndex
	ArrayIndex
	MapIndex
	Paren
	UnknownFieldAccess
	UnknownIndex
	SliceSlice
	ArraySlice
	UnknownSlice
	FuncCall
)

func GetExprMemAccKind(typeinfo *types.Info, expr ast.Expr) MemAccKind {
	switch expr := expr.(type) {
	case *ast.Ident:
		return Ident
	case *ast.StarExpr:
		return Deref
	case *ast.UnaryExpr:
		if expr.Op == token.AND {
			return Ref
		} else if expr.Op == token.MUL {
			return Deref
		}
	case *ast.SelectorExpr:
		typ := typeinfo.TypeOf(expr.X)
		if typ == nil {
			return UnknownFieldAccess
		}
		if _, ok := typ.Underlying().(*types.Pointer); ok {
			return FieldAccess
		} else if _, ok := typ.Underlying().(*types.Struct); ok {
			return IndirectFiledAccess
		} else {
			return UnknownFieldAccess
		}
	case *ast.IndexExpr:
		typ := typeinfo.TypeOf(expr.X)
		if typ == nil {
			return UnknownIndex
		}
		if _, ok := typ.Underlying().(*types.Slice); ok {
			return SliceIndex
		} else if _, ok := typ.Underlying().(*types.Array); ok {
			return ArrayIndex
		} else if _, ok := typ.Underlying().(*types.Map); ok {
			return MapIndex
		} else {
			return UnknownIndex
		}
	case *ast.SliceExpr:
		typ := typeinfo.TypeOf(expr.X)
		if typ == nil {
			return UnknownSlice
		}
		if _, ok := typ.Underlying().(*types.Slice); ok {
			return SliceSlice
		} else if _, ok := typ.Underlying().(*types.Array); ok {
			return ArraySlice
		} else {
			return UnknownSlice
		}
	case *ast.ParenExpr:
		return Paren
	default:
		return OtherMemAcc
	}
	return OtherMemAcc
}

// create table "ExprMemAcc" (
//
//	id bigint NOT NULL,
//	kind int NOT NULL,
//	"inner" bigint,
//	base_name text,
//	base_var integer,
//	depth integer NOT NULL,
//	depth_uncertain integer NOT NULL,
//	primary key(id),
//	foreign key(id) references "Expression"(id),
//	foreign key("inner") references "Expression"(id),
//	foreign key(base_var) references "Variable"(id),
//	check (depth>=0),
//	check (kind between 0 and 13),
//	check ( (base_var is NULL AND depth_uncertain>0)
//	    OR (base_var is NOT NULL AND depth_uncertain=0))
//
// );
func (conn *Connection) InsertExprMemAcc(exprID int64, kind MemAccKind, innerID int64, baseName string, baseVarID int64, depth int, depthUncertain int) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO "ExprMemAcc"(id,"inner",base_name,base_var,depth,depth_uncertain, kind) VALUES($1,$2,$3,$4,$5,$6,$7)`, exprID, to_null_int64(innerID), to_null_str(baseName), to_null_int64(baseVarID), depth, depthUncertain, kind)
	if err != nil {
		return err
	}
	return tx.Commit()
}
