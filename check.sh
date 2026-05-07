#!/bin/bash -x

function splitline {
    echo "-----------------"
}

# echo "Batch Clone Build Test"
# if ! go test -v ./cmd/batch_clone_build; then
#     echo "Batch Clone Build Test Failed"
#     exit 1
# fi

# splitline
# echo "Escape Adapter Test"
# if ! go test -v ./cmd/escape_adapter; then
#     echo "Escape Adapter Test Failed"
#     exit 1
# fi

splitline
echo "External Verify Test"

go run ./cmd/batch_clone_build yaml-examples/malloc_test.yaml
# check number of lines in pprof ext csv file is equal to count of CodeQL class instances
if ! go run ./cmd/pprof-external-verify codeql-db/test/malloc_test/ instance_count ; then
    echo "External Verify Test for malloc_test Failed"
    exit 1
fi

# check mallocgc time is mapped to CodeQL class instances
MALLOC_CSV="codeqlResult/pprof_ext/malloc_time/malloc_test.csv"
if [ ! -f "$MALLOC_CSV" ]; then
    echo "External Verify Test for mallocgc time mapping Failed: $MALLOC_CSV does not exist. Check if the query ran correctly."
    exit 1
fi
if [ "$(grep -c . "$MALLOC_CSV")" -ne 2 ]; then
    echo "External Verify Test for mallocgc time mapping Failed"
    echo "The content of mallocgc time csv file is:"
    cat "$MALLOC_CSV"
    exit 1
fi

echo "Done!"
