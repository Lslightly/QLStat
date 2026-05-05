package main

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lslightly/qlstat/utils"
	"github.com/stretchr/testify/assert"
)

func TestCloneBranch(t *testing.T) {
	var outbuf, errbuf bytes.Buffer
	if err := utils.RuncmdWithBuf(utils.ProjectRoot(), &outbuf, &errbuf,
		"go",
		"run",
		"./cmd/batch_clone_build",
		"-nobuild",
		"-noextgen",
		"./cmd/batch_clone_build/clone_test.yaml",
	); err != nil {
		assert.Nil(t, err, err.Error()+"\nstdout:"+outbuf.String()+"\nstderr:"+errbuf.String())
	}
	outbuf.Reset()
	assert.Nil(t, utils.RuncmdWithBuf(filepath.Join(utils.ProjectRoot(), "repos/github.com/zhihu-parallel-pi-calc"), &outbuf, nil, "git", "branch", "--show-current"))
	assert.Equal(t, "QLStat-test", strings.TrimSpace(outbuf.String()))
}
