package db

import (
	"database/sql"
	"go/ast"
	"go/token"
)
type StmtKind int

const (
	OtherStmt StmtKind = iota-1
	BadStmt
	DeclStmt
	EmptyStmt
	LabeledStmt
	ExprStmt
	SendStmt
	IncDecStmt
	AssignStmt
	GoStmt
	DeferStmt
	ReturnStmt
	BranchStmt
	BlockStmt
	IfStmt
	CaseClause
	SwitchStmt
	TypeSwitchStmt
	CommClause
	SelectStmt
	ForStmt
	RangeStmt
)

func GetStmtKind(stmt ast.Stmt) StmtKind {
	switch stmt.(type) {
	case *ast.DeclStmt:
		return DeclStmt
	case *ast.AssignStmt:
		return AssignStmt
	case *ast.ReturnStmt:
		return ReturnStmt
	case *ast.TypeSwitchStmt:
		return TypeSwitchStmt
	default:
		return OtherStmt
	}
}

func(conn *Connection) QueryStmt(pos, end token.Position) (int64, error) {
	file_id, err := conn.FileID(pos.Filename)
	if err != nil {
		return -1, err
	}
	return conn.queryStmt(file_id, pos.Line, pos.Column, end.Line, end.Column)
}

func(conn *Connection) queryStmt(file_id int64, line, column int, line_end, column_end int) (int64, error) {
	var rows *sql.Rows
	rows, err := conn.Query(`SELECT id FROM "Statement" WHERE file_id=$1 AND "line"=$2 AND "column"=$3 AND "line_end"=$4 AND "column_end"=$5`, file_id, line, column, line_end, column_end)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	for rows.Next() {
		var stmt_id int64
		err = rows.Scan(&stmt_id)
		if err != nil {
			return -1, nil
		}
		return stmt_id, nil
	}
	return -1, NotFoundError{}
}

func(conn *Connection) InsertStmt(kind StmtKind, pos,end token.Position, funcID, parentID int64) (int64, /*real inserted*/ bool, error) {
	filename := pos.Filename
	fileID, err := conn.FileID(filename)
	if err != nil {
		return -1, false, err
	}
	id, err := conn.queryStmt(fileID, pos.Line, pos.Column, end.Line, end.Column)
	if err==nil {
		return id, false, nil
	} else if _, ok := err.(NotFoundError); !ok {
		return -1, false, err
	}
	tx, err := conn.Begin()
	if err != nil {
		return -1, false, err
	}
	defer tx.Rollback()
	var rows *sql.Rows
	rows, err = tx.Query(`INSERT INTO "Statement"(kind,file_id,line,"column",line_end,column_end,func_id, parent_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, kind, fileID, pos.Line, pos.Column, end.Line, end.Column, to_null_int64(funcID), to_null_int64(parentID))
	if err != nil {
		return -1, false, err
	}
	if !rows.Next() {
		panic("no Statement id returned")
	}
	err = rows.Scan(&id)
	if err != nil {
		return -1, false, err
	}
	rows.Close()
	return id, true, tx.Commit()
}

// create table "StmtAssignLhs" (
//     id bigint NOT NULL,
//     pos int,
//     expr bigint NOT NULL,
//     primary key(id,pos),
//     foreign key(id) references "Statement"(id),
//     foreign key(expr) references "Expression"(id),
// 	check(pos>=0)
// );
func(conn *Connection) InsertStmtLhs(stmtID int64, pos int, exprID int64) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO "StmtAssignLhs"(id,pos,expr) VALUES($1,$2,$3)`, stmtID, pos, exprID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// create table "StmtAssignRhs" (
//     id bigint NOT NULL,
//     pos int,
//     expr bigint NOT NULL,
//     primary key(id,pos),
//     foreign key(id) references "Statement"(id),
//     foreign key(expr) references "Expression"(id),
// 	check(pos>=0)
// );
func(conn *Connection) InsertStmtRhs(stmtID int64, pos int, exprID int64) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO "StmtAssignRhs"(id,pos,expr) VALUES($1,$2,$3)`, stmtID, pos, exprID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// create table "StmtRetRes" (
//     id bigint NOT NULL,
//     pos int,
//     expr bigint NOT NULL,
//     primary key(id,pos),
//     foreign key(id) references "Statement"(id),
//     foreign key(expr) references "Expression"(id),
// 	check(pos>=0)
// );
func(conn *Connection) InsertStmtRetRes(stmtID int64, pos int, exprID int64) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO "StmtRetRes"(id,pos,expr) VALUES($1,$2,$3)`, stmtID, pos, exprID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// create table "StmtTypeSwitch" (
//     id bigint NOT NULL,
//     to_name text,
//     assert_expr bigint NOT NULL,
//     primary key(id),
//     foreign key(id) references "Statement"(id),
//     foreign key(assert_expr) references "Expression"(id)
// );
func(conn *Connection) InsertTypeSwitch(stmtID int64, name string, assertID int64) error {
	_, err := conn.Exec(`INSERT INTO "StmtTypeSwitch"(id,to_name,assert_expr) VALUES($1,$2,$3)`, stmtID, to_null_str(name), assertID)
	return err
}

func(conn *Connection) InsertTypeSwitchCase(stmtID int64, pos0, pos1 int, toTypeID int64) error {
	_, err := conn.Exec(`INSERT INTO "StmtTypeSwitchCase"(id,pos0,pos1,to_type) VALUES($1,$2,$3,$4)`, stmtID, pos0, pos1, to_null_int64(toTypeID))
	return err
}

// create table "StmtDefer" (
//     id bigint NOT NULL,
//     call_expr bigint NOT NULL,
//     primary key(id),
//     foreign key(id) references "Statement"(id),
//     foreign key(call_expr) references "Expression"(id)
// );
func(conn *Connection) InsertStmtDefer(stmtID int64, callExprID int64) error {
	_, err := conn.Exec(`INSERT INTO "StmtDefer"(id,call_expr) VALUES($1,$2)`, stmtID, callExprID)
	return err
}
