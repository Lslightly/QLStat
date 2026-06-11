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

func CreateQuery(path string, externals []string, externalFiles []string) Query {
	if filepath.Ext(path) != ".ql" {
		log.Fatalf("Suffix of query source %s is not .ql.", path)
	}
	cache := make([]string, len(externalFiles)*5)
	for _, extfile := range externalFiles {
		exts, err := ReadExtsFromFile(extfile)
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

// ExternalsSingleString returns ext1,ext2,...|files:extfile1,extfile2,...
func (q *Query) ExternalsSingleString() string {
	plainExts := strings.Join(q.externals, ",")
	bases := make([]string, 0, len(q.externalFiles))
	for _, extfile := range q.externalFiles {
		bases = append(bases, filepath.Base(extfile))
	}
	return plainExts + "|files:" + strings.Join(bases, ",")
}
