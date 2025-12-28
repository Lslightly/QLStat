package config

import (
	"log"
	"path/filepath"
	"strings"
)

type Query struct {
	path string
}

func CreateQuery(path string) Query {
	if filepath.Ext(path) != ".ql" {
		log.Fatalf("Suffix of query source %s is not .ql.", path)
	}
	return Query{
		path: path,
	}
}

func (q *Query) Name() string {
	return strings.TrimSuffix(filepath.Base(q.path), filepath.Ext(q.path))
}

func (q *Query) PathNoExt() string {
	return strings.TrimSuffix(q.path, filepath.Ext(q.path))
}

func (q *Query) AbsPathWithRoot(queryRoot string) string {
	return filepath.Join(queryRoot, q.path)
}

func (q *Query) AbsPathNoExtWithRoot(root string) string {
	return filepath.Join(root, q.PathNoExt())
}
