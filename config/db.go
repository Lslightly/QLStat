package config

import "path/filepath"

type DB struct {
	root string
	Name string
}

func (art *Artifact) DBCleanUpScriptPath() string {
	return filepath.Join(art.DBRoot, "cleanup_failed_directories.sh")
}

func (db *DB) Path() string {
	return filepath.Join(db.root, db.Name)
}

func (db *DB) ExtDir() string {
	return filepath.Join(db.Path(), "ext")
}
