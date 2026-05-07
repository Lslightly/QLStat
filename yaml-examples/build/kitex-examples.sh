#!/bin/bash
cd $REPO_DIR/hello
go build -a .
go build -a -o ./client-bin ./client


