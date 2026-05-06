#!/bin/bash -x

function splitline {
    echo "-----------------"
}

echo "Batch Clone Build Test"
if ! go test -v ./cmd/batch_clone_build; then
    echo "Batch Clone Build Test Failed"
    exit 1
fi

splitline
echo "Escape Adapter Test"
if ! go test -v ./cmd/escape_adapter; then
    echo "Escape Adapter Test Failed"
    exit 1
fi

splitline
echo "External Verify Test"

go run ./cmd/batch_clone_build -noclone yaml-examples/malloc_test.yaml
if ! go run ./cmd/pprof-external-verify codeql-db/test/malloc_test/ instance_count ; then
    echo "External Verify Test for malloc_test Failed"
    exit 1
fi

echo "Done!"
