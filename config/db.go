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
