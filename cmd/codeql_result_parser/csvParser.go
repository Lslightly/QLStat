// Copyright 2026 Qingwei Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
