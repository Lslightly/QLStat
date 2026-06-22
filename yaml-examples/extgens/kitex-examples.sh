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

cd $REPO_DIR/hello
> $OUTPUT_DIR/m2.log
go build -a -gcflags=all=-m=2 . 2>> $OUTPUT_DIR/m2.log
go build -a -gcflags=all=-m=2 -o ./client-bin ./client 2>> $OUTPUT_DIR/m2.log
cd $PROJROOT
go run ./cmd/escape_adapter -dir $DB_EXT_DIR -src=$REPO_DIR/hello -movedToHeap $OUTPUT_DIR/m2.log

