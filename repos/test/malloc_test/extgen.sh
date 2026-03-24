#!/bin/bash
ROOT=$(pwd)
PROJROOT=$ROOT/../../..
logdir=$PROJROOT/logs/extgen/test/malloc_test
go test -c -a -gcflags=-m=2 . &> $logdir/m2.log
cd $PROJROOT
go run ./cmd/escape_adapter -dir codeql-db/test/malloc_test/ext -src=$ROOT -movedToHeap -newEscapesToHeap $logdir/m2.log


