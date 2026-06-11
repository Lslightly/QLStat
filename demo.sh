#!/bin/bash
go run ./cmd/batch_clone_build ./demo.yaml # see logs/<clone/build/extgen/query> if you meet some errors
go run ./cmd/codeql_qdriver -collect ./demo.yaml
# Collect results to ./codeqlResult/<queryPathNoExt>.csv
