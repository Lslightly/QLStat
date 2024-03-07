package db

import (
	"database/sql"
	"go/token"
	"go/types"
	"log"
)

func(conn *Connection) QueryFunction(pos token.Position) (int64, error) {
	file_id, err := conn.FileID(pos.Filename)
	if err != nil {return 0, err}
	return conn.queryFunction(file_id, pos.Line, pos.Column)
}

func (conn *Connection) queryFunction(file_id int64, line int, column int) (int64, error) {
	rows, err := conn.Query(`SELECT id FROM "Function" WHERE file_id = $1 AND "line" = $2 AND "column" = $3`, file_id, line, column)
	if err != nil {return 0, err}
	defer rows.Close()
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		return id, err
	}
	return 0, NotFoundError{}
}

func (conn *Connection) InsertFunction(name string, pos token.Position, functype *types.Signature, is_literal bool, parent_func_id int64) (int64, error) {
	file_id, err := conn.FileID(pos.Filename)
	if err != nil {return 0, err}
	id, err := conn.queryFunction(file_id, pos.Line, pos.Column)
	if err!=nil {
		if _, ok := err.(NotFoundError); !ok {
			return 0, err
		}
	} else {
		return id, nil
	}

	// query type IDs
	var typeparam_types []int64
	var typeparam_names []sql.NullString
	recv_types := make([]int64, 0, 1)
	recv_names := make([]sql.NullString, 0, 1)
	param_types := make([]int64, 0, functype.Params().Len())
	param_names := make([]sql.NullString, 0, functype.Params().Len())
	result_types := make([]int64, 0, functype.Results().Len())
	result_names := make([]sql.NullString, 0, functype.Results().Len())
	if tp_params := functype.TypeParams(); tp_params != nil {
		for i:=0; i<tp_params.Len(); i++ {
			tp_param := tp_params.At(i)
			typeparam_names = append(typeparam_names, to_null_str(tp_param.String()))
			id, err := conn.TypeIDCheckNil(tp_param.Constraint())
			if err != nil {
				log.Printf("Error when query type parameter type %v: %v", tp_param.Constraint(), err)
			}
			typeparam_types = append(typeparam_types, id)
		}
	}
	if rec := functype.Recv(); rec != nil {
		id, err := conn.TypeIDCheckNil(rec.Type())
		if err != nil {
			log.Printf("InsertFunction: Error when query receiver type %v: %v", rec, err)
			id = -1
		}
		recv_types = append(recv_types, id)
		recv_names = append(recv_names, to_null_str(rec.Name()))
	}
	for i:=0; i<functype.Params().Len(); i++ {
		param := functype.Params().At(i)
		id, err := conn.TypeIDCheckNil(param.Type())
		if err != nil {
			log.Printf("InsertFunction: Error when query parameter type %v: %v", param, err)
			id = -1
		}
		param_names = append(param_names, to_null_str(param.Name()))
		param_types = append(param_types, id)
	}
	for i:=0; i<functype.Results().Len(); i++ {
		result := functype.Results().At(i)
		id, err := conn.TypeIDCheckNil(result.Type())
		if err != nil {
			log.Printf("InsertFunction: Error when query result type %v: %v", result, err)
			id = -1
		}
		result_names = append(result_names, to_null_str(result.Name()))
		result_types = append(result_types, id)
	}

	// var id int64
	tx, err := conn.Begin()
	if err != nil {return 0, err}
	defer tx.Rollback()
	rows, err := tx.Query(`INSERT INTO "Function"(name,is_variadic,file_id,line,"column",is_generic,is_literal,parent) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, to_null_str(name), functype.Variadic(), file_id, pos.Line, pos.Column, len(typeparam_names)!=0, is_literal, to_null_int64(parent_func_id))
	if err != nil {return 0, err}
	if !rows.Next() {panic("no id returned")}
	rows.Scan(&id)
	rows.Close()
	insert_helper := func(names []sql.NullString, types []int64, kind int) (err error) {
		for i := range names {
			_, err = tx.Exec(`INSERT INTO "FuncSig"(id,kind,pos,type,name) VALUES($1,$2,$3,$4,$5)`, id, kind, i, to_null_int64(types[i]), names[i])
			if err != nil {return}
		}
		return
	}
	err = insert_helper(recv_names, recv_types, 0)
	if err != nil {return 0, err}
	err = insert_helper(param_names, param_types, 1)
	if err != nil {return 0, err}
	err = insert_helper(result_names, result_types, 2)
	if err != nil {return 0, err}
	err = insert_helper(typeparam_names, typeparam_types, 3)
	if err != nil {return 0, err}
	tx.Commit()

	return id, nil
}
