#!/bin/bash
# codeql version 2.16.1
codeql database create -l=go -s . ../../codeql-db/pkgtest --command="go test -c ." --overwrite