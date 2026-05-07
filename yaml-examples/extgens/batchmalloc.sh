#!/bin/bash
cd $REPO_DIR
go build -a -gcflags=-m=2 . &> $OUTPUT_DIR/m2.log
cd $PROJROOT
go run ./cmd/escape_adapter -dir $DB_EXT_DIR -src=$REPO_DIR -movedToHeap -newEscapesToHeap $OUTPUT_DIR/m2.log

