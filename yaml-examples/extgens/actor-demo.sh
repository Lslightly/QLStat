#!/bin/bash
cd $PROJROOT
mkdir -p $DB_EXT_DIR
go run ./cmd/pprof2qlcsv -dir $DB_EXT_DIR $REPO_DIR/output/benchprofile/20260511-105838/BenchmarkFlowActor1.cpu.out
