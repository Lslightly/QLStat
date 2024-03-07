package db

import (
	"database/sql"
	"go/token"
	"go/types"
)

type FieldKind int
const (
	FieldOther FieldKind = iota
	FieldReceiver
	FieldParam
	FieldResult
)

// create table "Variable" (
//     id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
//     type bigint,
//     name text NOT NULL,
//     file_id integer NOT NULL,
//     parent_id bigint,
//     line integer NOT NULL,
//     "column" integer NOT NULL,
//     global boolean NOT NULL,
//     const boolean NOT NULL,
//     mutated integer,
//     primary key (id),
//     foreign key (type) references "Type" (id),
//     foreign key (file_id) references file(id),
//     foreign key (parent_id) references "Function"(id),
//     unique(file_id,line,"column"),
//     check(name <> '_'),
//     check(line >= 0),
//     check("column" >= 0),
//     check(mutated between 0 and 2)
// );
func (conn *Connection) QueryIdentDef(pos token.Position) (int64, error) {
	file_id, err := conn.FileID(pos.Filename)
	if err != nil {return 0, err}
	return conn.queryIdentDef(file_id, pos.Line, pos.Column)
}

func (conn *Connection) queryIdentDef(file_id int64, line int, column int) (int64, error) {
	rows, err := conn.Query(`SELECT id FROM "Variable" WHERE file_id = $1 AND "line" = $2 AND "column" = $3`, file_id, line, column)
	if err != nil {return 0, err}
	defer rows.Close()
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		return id, err
	}
	return 0, NotFoundError{}
}

func(conn *Connection) InsertIdentDef(name string, pos token.Position, func_id int64, parent_id int64, kind FieldKind, is_const bool, typeinfo types.Type) (int64, error) {
	file_id, err := conn.FileID(pos.Filename)
	if err != nil {return 0, err}
	id, err := conn.queryIdentDef(file_id, pos.Line, pos.Column)
	if err!=nil {
		if _, ok := err.(NotFoundError); !ok {
			return 0, err
		}
	} else {
		return id, nil
	}

	type_id, err := conn.TypeIDCheckNil(typeinfo)
	if err != nil {
		type_id = -1
	}
	tx, err := conn.Begin()
	if err != nil {return 0, err}
	defer tx.Rollback()
	// var id int64
	var rows *sql.Rows
	rows, err = tx.Query(`INSERT INTO "Variable"(type,name,file_id,line,"column",global,const, parent_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, to_null_int64(type_id), to_null_str(name), file_id, pos.Line, pos.Column, func_id==-1, is_const, to_null_int64(parent_id))
	if err != nil {return 0, err}

	if !rows.Next() {panic("no id returned")}
	rows.Scan(&id)
	rows.Close()
	// local
	if func_id != -1 {
		_, err := tx.Exec(`INSERT INTO "LocalVar"(id,param,return,receiver,func_id) VALUES($1,$2,$3,$4,$5)`, id, kind==FieldParam, kind==FieldResult, kind==FieldReceiver, func_id)
		if err != nil {return 0, err}
	} else {
		_, err := tx.Exec(`INSERT INTO "GlobalVar"(id) VALUES($1)`, id)
		if err != nil {return 0, err}
	}
	tx.Commit()
	return id, nil
}
