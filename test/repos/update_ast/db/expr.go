package db

import (
	"database/sql"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/barweiss/go-tuple"
)

type ExprKind int

const (
	BadExpr ExprKind = iota
	IdentExpr
	EllipsisExpr
	BasicLitExpr
	FuncLitExpr
	CompositeLitExpr
	ParenExpr
	SlectorExpr
	IndexExpr
	IndexListExpr
	SliceExpr
	TypeAssertExpr
	CallExpr
	StarExpr
	UnaryExpr
	BinaryExpr
	KeyValueExpr
	ArrayType
	StructType
	FuncType
	InterfaceType
	MapType
	ChanType
)

func GetExprKind(expr ast.Expr) ExprKind {
	switch expr.(type) {
		case *ast.Ident:
			return IdentExpr
		case *ast.Ellipsis:
			return EllipsisExpr
		case *ast.BasicLit:
			return BasicLitExpr
		case *ast.FuncLit:
			return FuncLitExpr
		case *ast.CompositeLit:
			return CompositeLitExpr
		case *ast.ParenExpr:
			return ParenExpr
		case *ast.SelectorExpr:
			return SlectorExpr
		case *ast.IndexExpr:
			return IndexExpr
		case *ast.IndexListExpr:
			return IndexListExpr
		case *ast.SliceExpr:
			return SliceExpr
		case *ast.TypeAssertExpr:
			return TypeAssertExpr
		case *ast.CallExpr:
			return CallExpr
		case *ast.StarExpr:
			return StarExpr
		case *ast.UnaryExpr:
			return UnaryExpr
		case *ast.BinaryExpr:
			return BinaryExpr
		case *ast.KeyValueExpr:
			return KeyValueExpr
		case *ast.ArrayType:
			return ArrayType
		case *ast.StructType:
			return StructType
		case *ast.FuncType:
			return FuncType
		case *ast.InterfaceType:
			return InterfaceType
		case *ast.MapType:
			return MapType
		case *ast.ChanType:
			return ChanType
		default:
			return BadExpr
	}
}

// -- 表达式
// create table "Expression" (
//     id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
//     type int,
//     kind int NOT NULL,
//     file_id integer NOT NULL,
//     "line" integer NOT NULL,
//     "column" integer NOT NULL,
//     "line_end" integer NOT NULL,
//     "column_end" integer NOT NULL,
//     func_id bigint,
//     parent_id bigint,
//     mem_acc integer,
//     primary key(id),
//     foreign key(file_id) references file(id),
//     foreign key(func_id) references "Function"(id),
//     foreign key(parent_id) references "Function"(id),
//     unique(file_id,"line","column","line_end","column_end"),
//     check ("line">=0),
//     check ("column">=0),
//     check ("line_end" >= "line"),
//     check ("column_end" >= "column"),
//     check (kind between 0 and 22),
//     check (mem_acc between 0 and 8)
// );
func(conn *Connection) InsertExpr(func_id, parent_id int64, pos,end token.Position,typ types.Type, kind ExprKind) (int64, /*real inserted*/  bool, error) {
	cacheID, err := conn.QueryExpr(pos,end)
	if err == nil {
		return cacheID, false, nil
	} else if _, ok := err.(NotFoundError); !ok{
		return -1, false, err
	}
	filename := pos.Filename
	fileID, err := conn.FileID(filename)
	if err != nil {
		return -1, false, err
	}
	typID, err := conn.TypeIDCheckNil(typ)
	if err != nil {
		typID = -1
	}
	var id int64
	var rows *sql.Rows
	rows, err = conn.Query(`INSERT INTO "Expression"(type,kind,file_id,"line","column",func_id,line_end,column_end,parent_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`, to_null_int64(typID), kind, fileID, pos.Line, pos.Column, to_null_int64(func_id), end.Line, end.Column, to_null_int64(parent_id))
	if err != nil {
		return -1, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		panic("no Expression id returned")
	}
	err = rows.Scan(&id)
	if err != nil {
		return -1, false, err
	}
	conn.expr_cache.Add(tuple.New2(pos,end), id)
	return id, true, nil
}

func(conn *Connection) QueryExpr(pos,end token.Position) (int64, error) {
	if id, ok := conn.expr_cache.Get(tuple.New2(pos,end)); ok {
		return id, nil
	}

	fileID, err := conn.FileID(pos.Filename)
	if err != nil {
		return -1, err
	}
	var rows *sql.Rows
	rows, err = conn.Query(`SELECT id FROM "Expression" WHERE file_id=$1 AND "line"=$2 AND "column"=$3 AND "line_end"=$4 AND "column_end"=$5`, fileID, pos.Line, pos.Column, end.Line, end.Column)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	for rows.Next() {
		var exprID int64
		err = rows.Scan(&exprID)
		if err != nil {
			return -1, nil
		}
		conn.expr_cache.Add(tuple.New2(pos,end), exprID)
		return exprID, nil
	}
	return -1, NotFoundError{}
}

// create table "ExprIdent" (
//     id bigint NOT NULL,
//     name text NOT NULL,
//     var_id integer,
//     primary key(id),
//     foreign key(id) references "Expression"(id),
//     foreign key(var_id) references "Variable"(id),
//     check (name <> '_')
// );
func(conn *Connection) InsertExprIdent(exprId int64, name string, obj_pos *token.Position) error {
	if obj_pos!=nil {
		if file_id, err := conn.FileID(obj_pos.Filename); err==nil {
			_, err := conn.Exec(`insert into "ExprIdent"(id,name,obj_file_id,obj_line,obj_column) VALUES($1,$2,$3,$4,$5)`, exprId, name, file_id, obj_pos.Line, obj_pos.Column)
			return err
		}
	}
	_, err := conn.Exec(`insert into "ExprIdent"(id,name) values($1,$2)`, exprId, name)
	return err
}

// create table "ExprSelector" (
//     id bigint NOT NULL,
//     base bigint NOT NULL,
//     field text NOT NULL,
//     primary key(id),
//     foreign key(id) references "Expression"(id),
//     foreign key(base) references "Expression"(id),
//     check (field <> '_')
// );
func(conn *Connection) InsertExprSelector(exprID int64, baseID int64, field string) error {
	// one statement is a transaction by itself, so we don't need to explicitly start a transaction
	// to execute this single statement
	_, err := conn.Exec(`INSERT INTO "ExprSelector"(id,base,field) VALUES($1,$2,$3)`, exprID, baseID, field)
	return err
}

// create table "ExprTypeAssert" (
//     id bigint NOT NULL,
//     base bigint NOT NULL,
//     to_type int,
//     primary key(id),
//     foreign key(id) references "Expression"(id),
//     foreign key(base) references "Expression"(id),
//     foreign key(to_type) references "Type"(id)
// );
func (conn *Connection) InsertExprTypeAssert(exprID int64, baseID int64, toTypeID int64) error {
	_, err := conn.Exec(`INSERT INTO "ExprTypeAssert"(id,base,to_type) VALUES($1,$2,$3)`, exprID, baseID, to_null_int64(toTypeID))
	return err
}

// create table "ExprCall" (
//     id bigint NOT NULL,
//     func bigint NOT NULL,
//     arg bigint NOT NULL,
//     pos int NOT NULL,
//     primary key(id,pos),
//     foreign key(id) references "Expression"(id),
//     foreign key(func) references "Expression"(id),
//     foreign key(arg) references "Expression"(id),
//     check (pos>=0)
// );
func (conn *Connection) InsertExprCall(exprID int64, funcID int64, argID int64, pos int) error {
	_, err := conn.Exec(`INSERT INTO "ExprCall"(id,func,arg,pos) VALUES($1,$2,$3,$4)`, exprID, funcID, argID, pos)
	return err
}
