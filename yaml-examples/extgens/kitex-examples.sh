#!/bin/bash
cd $REPO_DIR/hello
> $OUTPUT_DIR/m2.log
go build -a -gcflags=all=-m=2 . 2>> $OUTPUT_DIR/m2.log
go build -a -gcflags=all=-m=2 -o ./client-bin ./client 2>> $OUTPUT_DIR/m2.log
cd $PROJROOT
go run ./cmd/escape_adapter -dir $DB_EXT_DIR -src=$REPO_DIR/hello -movedToHeap $OUTPUT_DIR/m2.log

