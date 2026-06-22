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

package utils

import (
	"fmt"
	"io"
	"os/exec"
)

var Verbose bool = true

func Runcmd(dir string, name string, args ...string) error {
	return RuncmdWithBuf(dir, nil, nil, name, args...)
}

func RuncmdWithBuf(dir string, outbuf, errbuf io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if outbuf != nil {
		cmd.Stdout = outbuf
	}
	if errbuf != nil {
		cmd.Stderr = errbuf
	}
	cmd.Dir = dir
	if Verbose {
		fmt.Println(cmd.String())
	}
	return cmd.Run()
}
