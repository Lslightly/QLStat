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
