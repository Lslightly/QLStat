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

cd $REPO_DIR
go test -c -a -gcflags=-m=2 . &> $OUTPUT_DIR/m2.log
go test -run ^$ -bench . -cpuprofile $OUTPUT_DIR/cpu.out -memprofile $OUTPUT_DIR/mem.out &> $OUTPUT_DIR/bench.log
cd $PROJROOT
go run ./cmd/escape_adapter -dir $DB_EXT_DIR -src=$REPO_DIR -movedToHeap -newEscapesToHeap $OUTPUT_DIR/m2.log
go run ./cmd/pprof2qlcsv/ -dir $DB_EXT_DIR $OUTPUT_DIR/cpu.out

# queryLine for malloc written as external predicates
# it only applies to go1.24.2. Different go version may have different lines
cat <<EOF > $DB_EXT_DIR/queryLine.csv
runtime.mallocgcSmallScanNoHeader,1365,"span := c.alloc[spc]"
runtime.mallocgc,1060,"x, elemsize = mallocgcSmallScanNoHeader(size, typ, needzero)"
runtime.mallocgc,1049,"// Actually do the allocation."
EOF
