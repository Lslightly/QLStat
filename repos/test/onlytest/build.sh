#!/bin/bash
codeql database create -l=go -s . ../../../codeql-db/onlytest --command="go test -c ." --overwrite
