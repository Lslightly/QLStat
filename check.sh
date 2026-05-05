#!/bin/bash

function splitline {
    echo "-----------------"
}

echo "Batch Clone Build Test"
if ! go test ./cmd/batch_clone_build; then
    echo "Batch Clone Build Test Failed"
    exit 1
fi

splitline

echo "Escape Adapter Test"
if ! go test ./cmd/escape_adapter; then
    echo "Escape Adapter Test Failed"
    exit 1
fi

echo "Done!"
