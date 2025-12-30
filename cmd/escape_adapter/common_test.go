package main

import (
	"path/filepath"
	"runtime"
)

func curdir() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Dir(f)
}

func testdatadir() string {
	return filepath.Join(curdir(), "testdata")
}

func projectroot() string {
	return filepath.Dir(filepath.Dir(curdir()))
}

func codeqlResultDir() string {
	return filepath.Join(projectroot(), "codeqlResult")
}
