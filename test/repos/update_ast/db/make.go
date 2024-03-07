package db

import "fmt"

type Make struct {
	Fileid  int64
	Line    int
	Column  int
	Typestr string // slice, map, chan
	Cap     int64
}

func (m Make) print() {
	fmt.Println(m)
}

func (conn *Connection) InsertMake(row Make) {
	tx, err := conn.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	_, err = tx.Exec(`INSERT INTO "Make"(fileid,"line","column",type,cap) VALUES($1,$2,$3,$4,$5)`, row.Fileid, row.Line, row.Column, row.Typestr, row.Cap)
	if err != nil {
		fmt.Println(err)
		return
	}
	tx.Commit()
}
