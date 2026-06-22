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

package config

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

type Query struct {
	path                 string
	externals            []string
	externalFiles        []string
	cacheEntriesForFiles []string // cache external entries for externalFiles
}

func CreateQuery(path string, externals []string, externalFiles []string, externalRoot string) Query {
	if filepath.Ext(path) != ".ql" {
		log.Fatalf("Suffix of query source %s is not .ql.", path)
	}
	cache := make([]string, 0, len(externalFiles)*5)
	for _, extfile := range externalFiles {
		resolvedPath := extfile
		if !filepath.IsAbs(extfile) {
			resolvedPath = filepath.Join(externalRoot, extfile)
		}
		exts, err := ReadExtsFromFile(resolvedPath)
		if err != nil {
			log.Fatalf("Failed to read external file %s: %v", extfile, err)
		}
		cache = append(cache, exts...)
	}
	return Query{
		path:                 path,
		externals:            externals,
		externalFiles:        externalFiles,
		cacheEntriesForFiles: cache,
	}
}

func (q *Query) Name() string {
	return strings.TrimSuffix(filepath.Base(q.path), filepath.Ext(q.path))
}

func (q *Query) PathNoExt() string {
	return strings.TrimSuffix(q.path, filepath.Ext(q.path))
}

func (q *Query) QueryPath(queryRoot string) string {
	return filepath.Join(queryRoot, q.path)
}

func (q *Query) DirPath(root string) string {
	return filepath.Join(root, q.PathNoExt())
}

func (q *Query) ExternalOptions(extroot string) (res []string) {
	const format string = "--external=%s=%s"
	for _, ext := range q.externals {
		res = append(res, fmt.Sprintf(format, ext, filepath.Join(extroot, ext)+".csv"))
	}
	for _, ext := range q.cacheEntriesForFiles {
		res = append(res, fmt.Sprintf(format, ext, filepath.Join(extroot, ext)+".csv"))
	}
	return
}
