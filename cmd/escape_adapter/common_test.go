package main

import (
	"path/filepath"

	"github.com/Lslightly/qlstat/utils"
)

func testdatadir() string {
	return filepath.Join(utils.CurFileDir(), "testdata")
}
