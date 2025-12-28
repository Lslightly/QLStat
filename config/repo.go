package config

import (
	"log"
	"net/url"
	"os/exec"
	"path/filepath"
)

type Repo struct {
	FullName    string
	DirBaseName string
	*GitSource
}

func (r *Repo) RemoteURL() string {
	url, err := url.JoinPath(r.GitSource.Prefix, r.FullName+".git")
	if err != nil {
		log.Fatalf("Fail to know target url: %v", err)
	}
	return url
}

func (r *Repo) DirPath(root string) string {
	return filepath.Join(r.GitSource.HostDir(root), r.DirBaseName)
}

func (r *Repo) Clone(root string) error {
	cmd := exec.Command("git", "clone", r.RemoteURL(), r.DirPath(root))
	return cmd.Run()
}

func (r *Repo) DBPath(dbroot string) string {
	return filepath.Join(r.GitSource.HostDir(dbroot), r.DirBaseName)
}
