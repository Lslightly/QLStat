#!/bin/bash
ROOT=$(pwd)
PROJROOT=$ROOT/../../..
logdir=$PROJROOT/logs/extgen/test/batchmalloc
go build -a -gcflags=-m=2 . &> $logdir/m2.log
cd $PROJROOT
go run ./cmd/escape_adapter -dir codeql-db/test/batchmalloc/ext -src=$ROOT -movedToHeap -newEscapesToHeap $logdir/m2.log

