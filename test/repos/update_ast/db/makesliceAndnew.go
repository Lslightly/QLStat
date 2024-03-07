package db

import "fmt"

type Makeslice struct {
	Fileid   int64
	Line     int
	Column   int
	TypeStr  string // make([]T)'s []T
	TypeSize int64
	Size     int64 // size to allocate
}

type New struct {
	Fileid  int64
	Line    int
	Column  int
	TypeStr string // new(T)'s T
	Size    int64  // size to allocate
}

func (conn *Connection) InsertMakeslice(row Makeslice) {
	// fmt.Println(row)
	tx, err := conn.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO "Makeslice"(fileid,"line","column",type,typesize,size) VALUES($1,$2,$3,$4,$5,$6)`, row.Fileid, row.Line, row.Column, row.TypeStr, row.TypeSize, row.Size)
	if err != nil {
		fmt.Println(err)
		return
	}
	tx.Commit()
}

func (conn *Connection) InsertNew(row New) {
	// fmt.Println(row)
	tx, err := conn.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO "New"(fileid,"line","column",type,size) VALUES($1,$2,$3,$4,$5)`, row.Fileid, row.Line, row.Column, row.TypeStr, row.Size)
	if err != nil {
		fmt.Println(err)
		return
	}
	tx.Commit()
}
