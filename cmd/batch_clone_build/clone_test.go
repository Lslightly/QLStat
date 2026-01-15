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
	assert.Nil(t,
		utils.Runcmd(utils.ProjectRoot(),
			"go",
			"run",
			"./cmd/batch_clone_build",
			"-nobuild",
			"-noextgen",
			"./cmd/batch_clone_build/clone_test.yaml",
		),
	)
	var outbuf bytes.Buffer
	assert.Nil(t, utils.RuncmdWithBuf(filepath.Join(utils.ProjectRoot(), "repos/github.com/zhihu-parallel-pi-calc"), &outbuf, nil, "git", "branch", "--show-current"))
	assert.Equal(t, "QLStat-test", strings.TrimSpace(outbuf.String()))
}
