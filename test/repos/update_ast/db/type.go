package db

import (
	"database/sql"
	"fmt"
	"go/types"
	"log"
)

type Kind int64
const (
	KindInvalid Kind = iota
	KindBool
	KindInt
	KindFloat
	KindComplex
	KindArray
	KindPointer
	KindAddress
	KindSlice
	KindString
	KindFunction
	KindStruct
	KindInterface
	KindMap
	KindChannel
	KindNamed
	KindNil
	KindTypeParam
	KindTuple
	KindUnion
)

func(conn *Connection) TypeIDCheckNil(tp types.Type) (int64, error) {
	if tp == nil {return 0, TypeIsNil{}}
	return conn.TypeID(tp)
}

func(conn *Connection) TypeID(tp types.Type) (int64, error) {
	name := tp.String()
	// quick cache look-up
	if id,ok := func()(int64, bool) {
		conn.type_mutex.RLock()
		defer conn.type_mutex.RUnlock()
		if id, ok := conn.type_cache.Get(name); ok {
			return id, true
		}
		return 0, false
	}(); ok {
		return id, nil
	}

	conn.type_mutex.Lock()
	defer conn.type_mutex.Unlock()
	tx,err := conn.Begin()
	if err!=nil {return 0, err}
	defer tx.Rollback()
	id, err := conn.typeID(tp, tx)
	if err != nil {return 0, err}
	tx.Commit()
	return id, nil
}

func insertID(tx *sql.Tx, kind Kind, name string, length ...int64) (id int64, err error) {
	var rows *sql.Rows
	if len(length) ==0 {
		rows, err = tx.Query(`INSERT INTO "Type"(kind,name) VALUES($1,$2) RETURNING id`, kind, name)
	} else {
		rows, err = tx.Query(`INSERT INTO "Type"(kind,name,length) VALUES($1,$2,$3) RETURNING id`, kind, name, length[0])
	}
	if err != nil {
		// log.Printf("insertID(%s), fail", name)
		return
	}
	if !rows.Next() {return 0, fmt.Errorf("No id returned: rows.Next()=false")}
	err = rows.Scan(&id)
	if err != nil {return 0, err}
	err = rows.Close()
	// log.Printf("insertID(%s), id=%d", name, id)
	return
}

func(conn *Connection) typeID(tp types.Type, tx *sql.Tx) (int64, error) {
	// try fast looking-up by name first
	name := tp.String()
	if id, ok := conn.type_cache.Get(name); ok {
		return id, nil
	}

	rows, err := tx.Query(`SELECT id FROM "Type" WHERE name = $1`, name)
	if err!=nil {return 0,err}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		conn.type_cache.Add(name, id)
		// log.Printf("query(%s), id = %d", name, id)
		return id, nil
	}
	// log.Printf("query(%s), not found", name)

	// `tp` is not in DB
	var id int64
	switch tp := tp.(type) {
	case *types.Array: id,err = conn.typeIDArray(tp, tx, name)
	case *types.Basic: id,err = conn.typeIDBasic(tp, tx, name)
	case *types.Chan: id,err = conn.typeIDChan(tp, tx, name)
	case *types.Interface: id,err = conn.typeIDInterface(tp, tx, name)
	case *types.Map: id,err = conn.typeIDMap(tp, tx, name)
	case *types.Named: id,err = conn.typeIDNamed(tp, tx, name)
	case *types.Pointer: id,err = conn.typeIDPointer(tp, tx, name)
	case *types.Slice: id,err = conn.typeIDSlice(tp, tx, name)
	case *types.Struct: id,err = conn.typeIDStruct(tp, tx, name)
	case *types.Signature: id,err = conn.typeIDSignature(tp, tx, name)
	case *types.TypeParam: id,err = conn.typeIDTypeParam(tp, tx, name)
	case *types.Tuple: id,err = conn.typeIDTuple(tp, tx, name)
	case *types.Union: id,err = conn.typeIDUnion(tp, tx, name)
	default:
		return 0, fmt.Errorf("UNKNOWN type: %T; name = %s", tp, name)
	}
	if err != nil {
		log.Printf("typeID: ERROR while looking up %s: %v", name, err)
		return 0, err
	}
	if id==0 {
		panic(fmt.Sprintf("ID is zero: %s", name))
	}
	conn.type_cache.Add(name, id)
	return id, nil
}

func(conn *Connection) typeIDArray(tp *types.Array, tx *sql.Tx, name string) (int64, error) {
	size := tp.Len()
	base, err := conn.typeID(tp.Elem(), tx)
	if err != nil {return 0, err}
	id, err := insertID(tx, KindArray, name)
	if err != nil {return 0, err}
	_, err = tx.Exec(`INSERT INTO "TypeArray"(id,size,base) VALUES($1,$2,$3)`, id, size, base)
	return id, err
}

func(conn *Connection) typeIDBasic(tp *types.Basic, tx *sql.Tx, name string) (int64, error) {
	switch tp.Kind() {
	case types.Bool, types.UntypedBool:
		return insertID(tx, KindBool, name, 1)
	case types.Int:
		id, err := insertID(tx, KindInt, name, 8)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,TRUE)`, id)
		return id, err
	case types.Int8:
		id, err := insertID(tx, KindInt, name, 1)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,TRUE)`, id)
		return id, err
	case types.Int16:
		id, err := insertID(tx, KindInt, name, 2)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,TRUE)`, id)
		return id, err
	case types.Int32:
		id, err := insertID(tx, KindInt, name, 4)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,TRUE)`, id)
		return id, err
	case types.Int64:
		id, err := insertID(tx, KindInt, name, 8)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,TRUE)`, id)
		return id, err
	case types.Uint:
		id, err := insertID(tx, KindInt, name, 8)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,FALSE)`, id)
		return id, err
	case types.Uint8:
		id, err := insertID(tx, KindInt, name, 1)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,FALSE)`, id)
		return id, err
	case types.Uint16:
		id, err := insertID(tx, KindInt, name, 2)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,FALSE)`, id)
		return id, err
	case types.Uint32, types.UntypedRune:
		id, err := insertID(tx, KindInt, name, 4)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,FALSE)`, id)
		return id, err
	case types.Uint64:
		id, err := insertID(tx, KindInt, name, 8)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,FALSE)`, id)
		return id, err
	case types.Uintptr:
		id, err := insertID(tx, KindInt, name, 8)
		_, err = tx.Exec(`INSERT INTO "TypeInt"(id,signed) VALUES($1,FALSE)`, id)
		return id, err
	case types.Float32:
		return insertID(tx, KindFloat, name, 4)
	case types.Float64:
		return insertID(tx, KindFloat, name, 8)
	case types.Complex64:
		return insertID(tx, KindComplex, name, 8)
	case types.Complex128:
		return insertID(tx, KindComplex, name, 16)
	case types.String, types.UntypedString:
		return insertID(tx, KindString, name, 16)
	case types.UnsafePointer:
		return insertID(tx, KindAddress, name, 8)
	case types.Invalid:
		return insertID(tx, KindInvalid, name)
	case types.UntypedComplex:
		return insertID(tx, KindComplex, name)
	case types.UntypedFloat:
		return insertID(tx, KindFloat, name)
	case types.UntypedInt:
		return insertID(tx, KindInt, name)
	case types.UntypedNil:
		return insertID(tx, KindNil, name)
	default: panic("unreachable")
	}
}

func(conn *Connection) typeIDChan(tp *types.Chan, tx *sql.Tx, name string) (int64, error) {
	id, err := insertID(tx, KindChannel, name)
	if err != nil {return 0, err}
	direction := tp.Dir()
	base, err := conn.typeID(tp.Elem(), tx)
	if err!=nil {return 0, err}
	_, err = tx.Exec(`INSERT INTO "TypeChan"(id,direction,base) VALUES($1,$2,$3)`, id, direction, base)
	return id, err
}

func(conn *Connection) typeIDInterface(tp *types.Interface, tx *sql.Tx, name string) (int64, error) {
	id, err := insertID(tx, KindInterface, name)
	if err != nil {return 0, err}
	types := make([]int64, tp.NumMethods())
	// embedded := make([]bool, tp.NumMethods())
	names := make([]string, tp.NumMethods())
	for i:=0; i<tp.NumMethods(); i++ {
		f := tp.Method(i)
		// embedded[i] = f.Embedded()
		names[i] = f.Name()
		types[i], err = conn.typeID(f.Type(), tx)
		if err != nil {return 0, err}
	}

	for i := range types {
		_, err = tx.Exec(`INSERT INTO "TypeStructItfcField"(id,pos,name,type) VALUES($1,$2,$3,$4)`, id, i, names[i], types[i])
		if err != nil {return 0, err}
	}

	return id, nil
}

func(conn *Connection) typeIDMap(tp *types.Map, tx *sql.Tx, name string) (int64, error) {
	id, err := insertID(tx, KindMap, name)
	if err != nil {return 0, err}
	key, err := conn.typeID(tp.Key(), tx)
	if err!=nil {return 0, err}
	value, err := conn.typeID(tp.Elem(), tx)
	if err != nil {return 0, err}
	_, err = tx.Exec(`INSERT INTO "TypeMap"(id,key,value) VALUES($1,$2,$3)`, id, key, value)
	return id, err
}

func(conn *Connection) typeIDNamed(tp *types.Named, tx *sql.Tx, name string) (int64, error) {
	// name must be insert into Type table first, because the underlying type may
	// recursively refer to the name
	id, err := insertID(tx, KindNamed, name)
	if err != nil {return 0, err}

	// tp.TypeParams().At(0).Constraint()

	underlying, err := conn.typeID(tp.Underlying(), tx)
	if err != nil {return 0, err}
	local_name := tp.Obj().Name()
	var package_name string
	if tp.Obj().Pkg() != nil {
		package_name = tp.Obj().Pkg().Path()
		_, err = tx.Exec(`INSERT INTO "TypeNamed"(id,package,name,underlying) VALUES($1,$2,$3,$4)`, id, package_name, local_name, underlying)
	} else {
		_, err = tx.Exec(`INSERT INTO "TypeNamed"(id,name,underlying) VALUES($1,$2,$3)`, id, local_name, underlying)
	}
	return id, err
}

func(conn *Connection) typeIDPointer(tp *types.Pointer, tx *sql.Tx, name string) (int64, error) {
	id, err := insertID(tx, KindPointer, name)
	if err != nil {return 0, err}
	base, err := conn.typeID(tp.Elem(), tx)
	if err!=nil {return 0, err}
	_, err = tx.Exec(`INSERT INTO "TypePtrSlice"(id,base) VALUES($1,$2)`, id, base)
	return id, err
}

func(conn *Connection) typeIDSlice(tp *types.Slice, tx *sql.Tx, name string) (int64, error) {
	id, err := insertID(tx, KindSlice, name)
	if err != nil {return 0, err}
	base, err := conn.typeID(tp.Elem(), tx)
	if err!=nil {return 0, err}
	_, err = tx.Exec(`INSERT INTO "TypePtrSlice"(id,base) VALUES($1,$2)`, id, base)
	return id, err
}

func(conn *Connection) typeIDStruct(tp *types.Struct, tx *sql.Tx, name string) (int64, error) {
	id, err := insertID(tx, KindStruct, name)
	if err != nil {return 0, err}
	types := make([]int64, tp.NumFields())
	embedded := make([]bool, tp.NumFields())
	names := make([]string, tp.NumFields())
	for i:=0; i<tp.NumFields(); i++ {
		f := tp.Field(i)
		embedded[i] = f.Embedded()
		names[i] = f.Name()
		types[i], err = conn.typeID(f.Type(), tx)
		if err != nil {return 0, err}
	}

	for i := range types {
		if names[i] != "_" {
			_, err = tx.Exec(`INSERT INTO "TypeStructItfcField"(id,pos,embedding,name,type) VALUES($1,$2,$3,$4,$5)`, id, i, embedded[i], names[i], types[i])
		} else {
			_, err = tx.Exec(`INSERT INTO "TypeStructItfcField"(id,pos,embedding,type) VALUES($1,$2,$3,$4)`, id, i, embedded[i], types[i])
		}
		if err != nil {return 0, err}
	}

	return id, nil
}

func(conn *Connection) typeIDSignature(tp *types.Signature, tx *sql.Tx, name string) (int64, error) {
	id, err := insertID(tx, KindFunction, name)
	if err != nil {return 0, err}
	arg_types := make([]int64, tp.Params().Len())
	arg_names := make([]sql.NullString, tp.Params().Len())
	ret_types := make([]int64, tp.Results().Len())
	ret_names := make([]sql.NullString, tp.Results().Len())
	for i:=0; i<tp.Params().Len(); i++ {
		arg_types[i], err = conn.typeID(tp.Params().At(i).Type(), tx)
		arg_names[i] = to_null_str(tp.Params().At(i).Name())
		if err != nil {return 0, err}
	}
	for i:=0; i<tp.Results().Len(); i++ {
		ret_types[i], err = conn.typeID(tp.Results().At(i).Type(), tx)
		ret_names[i] = to_null_str(tp.Results().At(i).Name())
		if err != nil {return 0, err}
	}
	is_variadic := tp.Variadic()

	_, err = tx.Exec(`INSERT INTO "TypeFunc"(id,is_variadic,is_generic) VALUES($1,$2,$3)`, id, is_variadic, tp.TypeParams()!=nil)
	if err != nil {return 0, err}
	for i, arg := range arg_types {
		_, err = tx.Exec(`INSERT INTO "TypeFuncSig"(id,kind,pos,type,name) VALUES($1,1,$2,$3,$4)`, id, i, arg, arg_names[i])
		if err != nil {return 0, err}
	}
	for i, ret := range ret_types {
		_, err = tx.Exec(`INSERT INTO "TypeFuncSig"(id,kind,pos,type,name) VALUES($1,2,$2,$3,$4)`, id, i, ret, ret_names[i])
		if err != nil {return 0, err}
	}
	if tp.TypeParams() != nil {
		// typeparam_types := make([]int64, tp.TypeParams().Len())
		// typeparam_names := make([]sql.NullString, tp.TypeParams().Len())
		for i:=0; i<tp.TypeParams().Len(); i++ {
			name2 := to_null_str(tp.TypeParams().At(i).String())
			type2, err := conn.typeID(tp.TypeParams().At(i).Constraint(), tx)
			if err != nil {return 0, err}
			_, err = tx.Exec(`INSERT INTO "TypeFuncSig"(id,kind,pos,type,name) VALUES($1,3,$2,$3,$4)`, id, i, type2,name2)
		}
	}
	return id, err
}

func(conn *Connection) typeIDTypeParam(tp *types.TypeParam, tx *sql.Tx, name string) (int64, error) {
	return insertID(tx, KindTypeParam, name)
}

func(conn *Connection) typeIDTuple(tp *types.Tuple, tx *sql.Tx, name string) (int64, error) {
	return insertID(tx, KindTuple, name)
}

func(conn *Connection) typeIDUnion(tp *types.Union, tx *sql.Tx, name string) (int64, error) {
	id, err := insertID(tx, KindUnion, name)
	if err != nil {return 0, err}
	for i:=0; i<tp.Len(); i++ {
		term := tp.Term(i)
		term_type, err := conn.typeID(term.Type(), tx)
		if err != nil {return 0, err}
		_, err = tx.Exec(`INSERT INTO "TypeUnionTerm"(id,pos,may_underlying,type) VALUES($1,$2,$3,$4)`, id, i, term.Tilde(), term_type)
		if err != nil {return 0, err}
	}
	return id, nil
}
