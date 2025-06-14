package main

import (
	"encoding/csv"
	"errors"
	"os"
	"path"
	"strings"
)

type CodeQLCSV struct {
	data   [][]string
	path   string // the path of csv, like `qlName/repoName.csv`
	qlName string // name of codeql query
}

func NewCSV(path string, qlName string) (res CodeQLCSV, err error) {
	res.path = path
	res.qlName = qlName

	infile, err := os.Open(path)
	if err != nil {
		return
	}

	r := csv.NewReader(infile)
	res.data, err = r.ReadAll()
	if err != nil {
		return
	}

	if len(res.data) < 1 { // check validation of csv file
		err = errors.New(path + "has < 2 lines")
		return
	}

	return res, nil
}

func (this CodeQLCSV) GetRepoName() string {
	return strings.TrimSuffix(path.Base(this.path), ".csv")
}
