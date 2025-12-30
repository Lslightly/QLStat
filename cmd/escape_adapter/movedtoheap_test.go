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

func movedToHeapSetup(t *testing.T) {
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
}

// need more timeout
func TestCodeQLMovedToHeap1(t *testing.T) {
	movedToHeapSetup(t)
	moved_to_heap_var_test(t)
	inlined_var_test(t)
}

func TestCodeQLMovedToHeap2(t *testing.T) {
	movedToHeapSetup(t)
	ref_in_go_test(t)
	heapvar_use_in_go_test(t)
}

func moved_to_heap_var_test(t *testing.T) {
	csvPath := filepath.Join(codeqlResultDir(), "escape_ext/moved_to_heap_var_test/escape.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 12)
}

func inlined_var_test(t *testing.T) {
	csvPath := filepath.Join(codeqlResultDir(), "escape_ext/inlined_var_test/false-sharing.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 1)
}

func ref_in_go_test(t *testing.T) {
	csvPath := filepath.Join(codeqlResultDir(), "escape_ext/ref_in_go_test/false-sharing.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 28)
}

func heapvar_use_in_go_test(t *testing.T) {
	csvPath := filepath.Join(codeqlResultDir(), "escape_ext/heapvar_use_in_go_test/false-sharing.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 4)
}
