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
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func convint(s string) int {
	res, err := strconv.Atoi(s)
	if err != nil {
		log.Panicf("error converting %s to int: %v", s, err)
	}
	return res
}

func cleanpath(path string) string {
	absSrcRoot, err := filepath.Abs(SrcRoot)
	if err != nil {
		log.Panicf("error when converting SrcRoot: %s: %v", SrcRoot, err)
	}
	path = filepath.Clean(path)
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Clean(filepath.Join(absSrcRoot, path))
}

type CSVRow struct {
	path                string
	startLine, startCol int
}

func (row *CSVRow) String() string {
	const pat string = "%s,%d,%d"
	return fmt.Sprintf(pat, row.path, row.startLine, row.startCol)
}

// path, startLine, startCol
func movedToHeapHandle(lineGen LineGenerator) (csvRows []string) {
	rowSet := make(map[CSVRow]bool)
	for i, line := range lineGen {
		if !strings.Contains(line, "moved to heap") {
			continue
		}
		regex := regexp.MustCompile(`(.*?):(\d+):(\d+):`)
		matches := regex.FindStringSubmatch(line)
		if len(matches) == 0 {
			log.Printf("line %d with moved to heap but no match\n", i+1)
			continue
		}
		path, startLineStr, startColStr := matches[1], matches[2], matches[3]
		row := CSVRow{
			path:      cleanpath(path),
			startLine: convint(startLineStr),
			startCol:  convint(startColStr),
		}
		if rowSet[row] {
			continue
		}
		rowSet[row] = true
	}
	csvRows = make([]string, 0, len(rowSet))
	for row := range rowSet {
		csvRows = append(csvRows, row.String())
	}
	return
}

func newEscapesToHeapHandle(lineGen LineGenerator) (csvRows []string) {
	rowSet := make(map[CSVRow]bool)
	for i, line := range lineGen {
		if !(strings.Contains(line, "new") && strings.Contains(line, "escapes to heap")) { // does not contain (new && escapes to heap)
			continue
		}
		regex := regexp.MustCompile(`(.*?):(\d+):(\d+):`)
		matches := regex.FindStringSubmatch(line)
		if len(matches) == 0 {
			log.Printf("line %d with moved to heap but no match\n", i+1)
			continue
		}
		if len(regexp.MustCompile(`new\(.*\) escapes to heap`).FindStringSubmatch(line)) == 0 { // does not match new(.*?) escapes to heap
			log.Printf("line %d does not match new\\(.*?\\) escapes to heap", i+1)
			continue
		}
		path, startLineStr, startColStr := matches[1], matches[2], matches[3]
		row := CSVRow{
			path:      cleanpath(path),
			startLine: convint(startLineStr),
			startCol:  convint(startColStr),
		}
		if rowSet[row] {
			continue
		}
		rowSet[row] = true
	}
	csvRows = make([]string, 0, len(rowSet))
	for row := range rowSet {
		csvRows = append(csvRows, row.String())
	}
	return
}
