#!/bin/bash
go run ./cmd/batch_clone_build ./test.yaml # see logs/<clone/build/extgen/query> if you meet some errors
go run ./cmd/codeql_qdriver -collect ./test.yaml
# Results are in ./codeqlResult/escape_ext/heapvar_should_move
