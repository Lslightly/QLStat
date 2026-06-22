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

const (
	REPO_DIR   = "REPO_DIR"   // the root directory of the repository
	OUTPUT_DIR = "OUTPUT_DIR" // the directory to store intermediate results for generating external predicate
	PROJROOT   = "PROJROOT"   // the root directory of the project
	DB_EXT_DIR = "DB_EXT_DIR" // the directory to store external predicate database
)

type envpair struct {
	name, value string
}

// genEnv converts envpairs to strings in the format of "name=value". This format is required by cmd.Env in exec.Command.
func genEnv(pairs []envpair) (res []string) {
	for _, pair := range pairs {
		res = append(res, pair.name+"="+pair.value)
	}
	return res
}

func allAbs(pairs []envpair) (res []envpair) {
	for _, pair := range pairs {
		res = append(res, envpair{
			name:  pair.name,
			value: abspath(pair.value),
		})
	}
	return res
}
