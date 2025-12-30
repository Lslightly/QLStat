package config

import "path/filepath"

func (art *Artifact) DBCleanUpScriptPath() string {
	return filepath.Join(art.DBRoot, "cleanup_failed_directories.sh")
}
