package db

import (
	"fmt"
	"strings"
)

// Set this variable to config.RepoDir before analyzing.
var RepoDir string = "[uninitialized]"
var repoDirExpected string = "/data/github_go/repos"

func fix_path(path string) string {
	if RepoDir != repoDirExpected && strings.HasPrefix(path, RepoDir+"/") {
		return repoDirExpected + "/" + strings.TrimPrefix(path, RepoDir+"/")
	} else {
		return path
	}
}

// Get repository ID in the database. The repository information must be pre-stored in the database.
// - name: name of the repository
func (conn *Connection) RepoID(name string) (id int64, err error) {
	rows, err := conn.Query(`SELECT id FROM "repository" WHERE repo_name = $1`, name)
	if err != nil {return}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&id)
		return
	}
	return 0, fmt.Errorf("RepoID(%s): repository not found, please run undata_repo_info first", name)
}

// Get package id by path.
// - directory: directory of the package in the filesystem.
// - name: package name
func (conn *Connection) PackageID(directory string, name string) (id int64, err error) {
	directory = fix_path(directory)
	rows, err := conn.Query(`SELECT id FROM "package" WHERE path = $1 AND name = $2`, directory, name)
	if err != nil {return}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&id)
		return
	}
	return 0, fmt.Errorf("PackageID(%s): package not found, please run update_pkg_file first", directory)
}

// Get package id by path.
// - path: import path of the package.
func (conn *Connection) PackageIDByImport(path string) (id int64, err error) {
	path = fix_path(path)
	rows, err := conn.Query(`SELECT id FROM "file" WHERE import_path = $1`, path)
	if err != nil {return}
	defer rows.Close()
	ids := make([]int64,0,1)
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {return}
		ids = append(ids, id)
	}
	if len(ids) == 1{
		return ids[0], nil
	} else if len(ids) == 0{
		return 0, fmt.Errorf("PackageIDByImport(%s): package not found, please run update_pkg_file first", path)
	} else {
		return 0, fmt.Errorf("PackageIDByImport(%s): ambiguous import path, found packages %v", path, ids)
	}
}

// Get file ID by path.
// - path: path of the file in the filesystem.
func (conn *Connection) FileID(path string) (id int64, err error) {
	path = fix_path(path)
	if id, ok := conn.file_cache.Get(path); ok {
		return id, nil
	}
	rows, err := conn.Query(`SELECT id FROM "file" WHERE path = $1`, path)
	if err != nil {return}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&id)
		if err == nil {
			conn.file_cache.Add(path, id)
		}
		return
	}
	return 0, fmt.Errorf("FileID(%s): file not found, please run update_pkg_file first", path)
}
