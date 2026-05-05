#!/bin/bash

function splitline {
    echo "-----------------"
}

echo "Batch Clone Build Test"
go test ./cmd/batch_clone_build

splitline

echo "Escape Adapter Test"
go test ./cmd/escape_adapter

echo "Done!"
