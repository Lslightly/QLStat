package main

import (
	"encoding/csv"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func movedtoheapYaml() string {
	return filepath.Join(curdir(), "movedtoheap_test.yaml")
}

func TestMovedToHeap(t *testing.T) {
	lines := readLinesFromFile(filepath.Join(testdatadir(), "escape.log"))
	rows := movedToHeapHandle(lines)

	assert.Len(t, rows, 28)
}

func runcmd(name string, args []string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func TestCodeQLMovedToHeap(t *testing.T) {
	os.Chdir(projectroot())
	assert.Nil(
		t,
		runcmd("go", []string{
			"run",
			"./cmd/batch_clone_build",
			"-noclone",
			movedtoheapYaml(),
		}),
	)
	assert.Nil(
		t,
		runcmd("go", []string{
			"run",
			"./cmd/codeql_qdriver",
			"-collect",
			movedtoheapYaml(),
		}),
	)
	csvPath := filepath.Join(codeqlResultDir(), "escape_ext/moved_to_heap_var_test/escape.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 12)

	csvPath = filepath.Join(codeqlResultDir(), "escape_ext/inlined_var_test/false-sharing.csv")
	f, err = os.Open(csvPath)
	assert.Nil(t, err)
	reader = csv.NewReader(f)
	recs, err = reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 1)
}
