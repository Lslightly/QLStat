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
	"log"
	"os"
	"path/filepath"
	"runtime"
)

func CreateOutAndErr(sharedPathNoExt string) (outFile, errFile *os.File) {
	outFile = CreateFile(sharedPathNoExt + ".out")
	errFile = CreateFile(sharedPathNoExt + ".err")
	return
}

func MkdirAll(dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Panicf("error when creating dir %s: %v", dir, err)
	}
}

func CreateFile(file string) *os.File {
	dirname := filepath.Dir(file)
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		MkdirAll(dirname)
	}
	res, err := os.Create(file)
	if err != nil {
		log.Panicf("error when creating file %s: %v", res.Name(), err)
	}
	return res
}

func Remkdir(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Panicf("error occurs when deleting dir %s: %v", dir, err)
	}
	MkdirAll(dir)
}

func OpenFile(f string) *os.File {
	res, err := os.Open(f)
	if err != nil {
		log.Panicf("error occurs when open file %s: %v", res.Name(), err)
	}
	return res
}

func CurFileDir() string {
	_, file, _, _ := runtime.Caller(1)
	return filepath.Dir(file)
}

func ProjectRoot() string {
	return filepath.Dir(CurFileDir())
}
