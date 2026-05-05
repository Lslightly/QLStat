#!/bin/bash
ROOT=$(pwd)
PROJROOT=$ROOT/../../..
logdir=$PROJROOT/logs/extgen/test/malloc_test
go test -c -a -gcflags=-m=2 . &> $logdir/m2.log
go test -run ^$ -bench . -cpuprofile $logdir/cpu.out &> $logdir/bench.log
cd $PROJROOT
go run ./cmd/escape_adapter -dir codeql-db/test/malloc_test/ext -src=$ROOT -movedToHeap -newEscapesToHeap $logdir/m2.log
go run ./cmd/pprof2qlcsv/ -dir codeql-db/test/malloc_test/ext $logdir/cpu.out
