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
	*RepoGroup
}

// HasURLPrefix returns true if the repo belongs to a group with a URL prefix (remote repo).
func (r *Repo) HasURLPrefix() bool {
	return r.RepoGroup != nil && r.RepoGroup.URLPrefix != ""
}

// RemoteURL returns the full clone URL by joining URLPrefix and FullName.
// Must only be called when HasURLPrefix() is true.
func (r *Repo) RemoteURL() string {
	if !r.HasURLPrefix() {
		log.Fatalf("RemoteURL called on repo %s which has no URLPrefix (local-only repo)", r.FullName)
	}
	u, err := url.JoinPath(r.RepoGroup.URLPrefix, r.FullName)
	if err != nil {
		log.Fatalf("Fail to know target url: %v", err)
	}
	u += ".git"
	return u
}

// DirPath returns root/<Dir>/<dirBaseName> if Dir is set, otherwise root/<dirBaseName>
func (r *Repo) DirPath(root string) string {
	return filepath.Join(r.RepoGroup.GroupDir(root), r.DirBaseName)
}

// Clone clones the repo from remote. Skips silently if URLPrefix is empty (local-only repo).
func (r *Repo) Clone(root string) error {
	if !r.HasURLPrefix() {
		log.Printf("Skipping clone for %s: no URLPrefix (local-only repo)", r.FullName)
		return nil
	}
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
	return filepath.Join(dbroot, r.DirBaseName)
}

func (r *Repo) DBExtDir(dbroot string) string {
	return filepath.Join(r.DBPath(dbroot), "ext")
}

func (r *Repo) ExtGenDir(dbroot string) string {
	return filepath.Join(r.DBPath(dbroot), "extgen")
}
