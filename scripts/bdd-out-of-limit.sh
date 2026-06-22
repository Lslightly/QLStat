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

## Script for issue https://github.com/github/codeql/issues/21961
# Execute the script in the project root directory

# The steps in the shell script are internal steps of two commands. (You don't need care these two commands actually)
# 1. go run ./cmd/batch_clone_build -noclone yaml-examples/malloc_test.yaml
# 2. go run ./cmd/codeql_qdriver -collect yaml-examples/malloc_test.yaml

# important steps in 1. `go run ./cmd/batch_clone_build -noclone yaml-examples/malloc_test.yaml`
mkdir -p codeql-db/malloc_test/extgen codeql-db/malloc_test/ext
export REPO_DIR=$(realpath repos/test/malloc_test)
export OUTPUT_DIR=$(realpath codeql-db/malloc_test/extgen)
export PROJROOT=$(realpath "$(dirname "$0")/..")
export DB_EXT_DIR=$(realpath codeql-db/malloc_test/ext)
## build
codeql database create codeql-db/malloc_test -l=go --overwrite -s=repos/test/malloc_test -c $PROJROOT/yaml-examples/build/malloc_test.sh
## extgen
mkdir -p codeql-db/malloc_test/extgen codeql-db/malloc_test/ext
cd $REPO_DIR
go test -run ^$ -bench . -cpuprofile $OUTPUT_DIR/cpu.out -memprofile $OUTPUT_DIR/mem.out &> $OUTPUT_DIR/bench.log
cd $PROJROOT
go run ./cmd/pprof2qlcsv/ -dir $DB_EXT_DIR $OUTPUT_DIR/mem.out
# now in codeql-db/malloc_test/ext, there are csv tables of pprof heap profile data


# important steps in 2. `go run ./cmd/codeql_qdriver -collect yaml-examples/malloc_test.yaml`
## query
mkdir -p codeqlResult/pprof_ext/heap_test
codeql query run -d=codeql-db/malloc_test --search-path=qlsrc/lib qlsrc/pprof_ext/heap_test.ql --output=codeqlResult/pprof_ext/heap_test/malloc_test.bqrs --external=profile=codeql-db/malloc_test/ext/profile.csv --external=value_type=codeql-db/malloc_test/ext/value_type.csv --external=sample=codeql-db/malloc_test/ext/sample.csv --external=sample_to_location_id=codeql-db/malloc_test/ext/sample_to_location_id.csv --external=sample_to_value=codeql-db/malloc_test/ext/sample_to_value.csv --external=sample_to_label=codeql-db/malloc_test/ext/sample_to_label.csv --external=label=codeql-db/malloc_test/ext/label.csv --external=mapping=codeql-db/malloc_test/ext/mapping.csv --external=location=codeql-db/malloc_test/ext/location.csv --external=location_to_line=codeql-db/malloc_test/ext/location_to_line.csv --external=line=codeql-db/malloc_test/ext/line.csv --external=function=codeql-db/malloc_test/ext/function.csv --external=string_table=codeql-db/malloc_test/ext/string_table.csv --external=profile_to_sample_type=codeql-db/malloc_test/ext/profile_to_sample_type.csv --external=profile_to_sample=codeql-db/malloc_test/ext/profile_to_sample.csv --external=profile_to_mapping=codeql-db/malloc_test/ext/profile_to_mapping.csv --external=profile_to_location=codeql-db/malloc_test/ext/profile_to_location.csv --external=profile_to_function=codeql-db/malloc_test/ext/profile_to_function.csv --external=profile_to_string_table=codeql-db/malloc_test/ext/profile_to_string_table.csv --external=profile_to_comment=codeql-db/malloc_test/ext/profile_to_comment.csv


