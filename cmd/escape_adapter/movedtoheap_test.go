package main

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func curdir() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Dir(f)
}

func testdatadir() string {
	return filepath.Join(curdir(), "testdata")
}

func TestMovedToHeap(t *testing.T) {
	lines := readLinesFromFile(filepath.Join(testdatadir(), "escape.log"))
	rows := movedToHeapHandle(lines)

	assert.Len(t, rows, 28)
}
