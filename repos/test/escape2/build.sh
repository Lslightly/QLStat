#!/bin/bash
# Copyright 2026 Qingwei Li
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

go list -export -f '{{if .Export}}packagefile {{.ImportPath}}={{.Export}}{{end}}' std > importcfg
go tool compile -d=panic -p=p -C -e -importcfg=./importcfg -o a.o -l -d=ssa/check/on -m=2 escape2.go &> m2.log
