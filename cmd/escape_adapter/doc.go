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

/*
escape analysis adapter

The adapter will adapt the escape analysis -m=2 output to CSV files
and store these files in $dbRoot/<repo>/ext/<pred>.csv. Then
these adapted results can be used by `codeql query run --external`
to extend the ability of CodeQL.

The input of the adapter is:
1. the predicate to generate
2. the output csv dir
3. the input escape analysis log

The format of output csv file should follow ext_preds.qll in
qlsrc/escape_ext/ext_preds.qll.
*/
