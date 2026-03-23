#!/bin/bash
codeql database create -l=go -s . ../../../codeql-db/test/onlytest --command="go test -c ." --overwrite
