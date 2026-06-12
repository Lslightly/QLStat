#!/bin/bash
mkdir -p bin
go build -o ./bin/batch_clone_build ./cmd/batch_clone_build
go build -o ./bin/codeql_qdriver ./cmd/codeql_qdriver
