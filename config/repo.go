package config

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"path/filepath"

	"github.com/Lslightly/qlstat/utils"
)

type Repo struct {
	FullName    string
	DirBaseName string
	branch      string
	*GitSource
}

func (r *Repo) RemoteURL() string {
	url, err := url.JoinPath(r.GitSource.Prefix, r.FullName)
	if err != nil {
		log.Fatalf("Fail to know target url: %v", err)
	}
	url += ".git"
	return url
}

func (r *Repo) DirPath(root string) string {
	return filepath.Join(r.GitSource.HostDir(root), r.DirBaseName)
}

func (r *Repo) Clone(root string) error {
	args := []string{"clone"}
	args = append(args, r.RemoteURL(), r.DirPath(root))
	var errBuf bytes.Buffer
	err := utils.RuncmdWithBuf(utils.ProjectRoot(), nil, &errBuf, "git", args...)
	fmt.Println(errBuf.String())
	return err
}

func (r *Repo) Checkout(root string) error {
	if r.branch == "" {
		return fmt.Errorf("checkout must have branch specified. repo %s does not have branch specified.", r.FullName)
	}
	return utils.Runcmd(r.DirPath(root), "git", "checkout", r.branch)
}

func (r *Repo) DBPath(dbroot string) string {
	return filepath.Join(r.GitSource.HostDir(dbroot), r.DirBaseName)
}

func (r *Repo) DBExtDir(dbroot string) string {
	return filepath.Join(r.DBPath(dbroot), "ext")
}
