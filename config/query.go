package config

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

type Query struct {
	path      string
	externals []string
}

func CreateQuery(path string, externals []string) Query {
	if filepath.Ext(path) != ".ql" {
		log.Fatalf("Suffix of query source %s is not .ql.", path)
	}
	return Query{
		path:      path,
		externals: externals,
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
	return
}
