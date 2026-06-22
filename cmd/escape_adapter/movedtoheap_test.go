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
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Lslightly/qlstat/utils"
	"github.com/stretchr/testify/assert"
)

func movedtoheapYaml() string {
	return filepath.Join(utils.CurFileDir(), "movedtoheap_test.yaml")
}

var (
	once sync.Once
)

func TestMovedToHeap(t *testing.T) {
	rows := movedToHeapHandle(createLineGen(filepath.Join(testdatadir(), "escape.log")))

	assert.Len(t, rows, 28)
}

func movedToHeapSetup(t assert.TestingT) {
	assert.Nil(
		t,
		utils.Runcmd(utils.ProjectRoot(), "go", []string{
			"run",
			"./cmd/batch_clone_build",
			"-noclone",
			movedtoheapYaml(),
		}...),
	)
	assert.Nil(
		t,
		utils.Runcmd(utils.ProjectRoot(), "go", []string{
			"run",
			"./cmd/codeql_qdriver",
			"-collect",
			movedtoheapYaml(),
		}...),
	)
}

// need more timeout
func TestCodeQLMovedToHeap1(t *testing.T) {
	once.Do(func() {
		movedToHeapSetup(t)
	})
	moved_to_heap_var_test(t)
	inlined_var_test(t)
}

func TestCodeQLMovedToHeap2(t *testing.T) {
	once.Do(func() {
		movedToHeapSetup(t)
	})
	ref_in_go_test(t)
	heapvar_use_in_go_test(t)
}

func TestCodeQLMovedToHeap3(t *testing.T) {
	once.Do(func() {
		movedToHeapSetup(t)
	})
	same_scope_go_ref_heapvar_test(t)
}

func moved_to_heap_var_test(t *testing.T) {
	csvPath := filepath.Join(utils.ProjectRoot(), "codeql-db/escape/results/lslightly/qlstat/escape_ext/moved_to_heap_var_test.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 12)
}

func inlined_var_test(t *testing.T) {
	csvPath := filepath.Join(utils.ProjectRoot(), "codeql-db/false-sharing/results/lslightly/qlstat/escape_ext/inlined_var_test.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 1)
}

func ref_in_go_test(t *testing.T) {
	csvPath := filepath.Join(utils.ProjectRoot(), "codeql-db/false-sharing/results/lslightly/qlstat/escape_ext/ref_in_go_test.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 28)
}

func heapvar_use_in_go_test(t *testing.T) {
	csvPath := filepath.Join(utils.ProjectRoot(), "codeql-db/false-sharing/results/lslightly/qlstat/escape_ext/heapvar_use_in_go_test.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 4)
}

func same_scope_go_ref_heapvar_test(t *testing.T) {
	csvPath := filepath.Join(utils.ProjectRoot(), "codeql-db/false-sharing/results/lslightly/qlstat/escape_ext/same_scope_go_ref_heapvar_test.csv")
	f, err := os.Open(csvPath)
	assert.Nil(t, err)
	reader := csv.NewReader(f)
	recs, err := reader.ReadAll()
	assert.Nil(t, err)
	recs = recs[1:] // remove header
	assert.Len(t, recs, 1)
}
