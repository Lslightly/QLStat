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

go run ./cmd/batch_clone_build yaml-examples/malloc_test.yaml
# check number of lines in pprof ext csv file is equal to count of CodeQL class instances
if ! go run ./cmd/pprof-external-verify codeql-db/test/malloc_test/ instance_count ; then
    echo "External Verify Test for malloc_test Failed"
    exit 1
fi

go run ./cmd/codeql_qdriver -collect yaml-examples/malloc_test.yaml
# check mallocgc time is mapped to CodeQL class instances
malloc_time_csv="codeqlResult/pprof_ext/malloc_time/malloc_test.csv"
if [ ! -f "$malloc_time_csv" ]; then
    echo "External Verify Test for mallocgc time mapping Failed: $malloc_time_csv does not exist. Check if the query ran correctly."
    exit 1
fi
if [ "$(grep -c . "$malloc_time_csv")" -ne 2 ]; then
    echo "External Verify Test for mallocgc time mapping Failed"
    echo "The content of mallocgc time csv file is:"
    cat "$malloc_time_csv"
    exit 1
fi

# check malloc_line has results
malloc_line_csv="codeqlResult/pprof_ext/malloc_line/malloc_test.csv"
if [ ! -f "$malloc_line_csv" ]; then
    echo "External Verify Test for mallocgc time mapping Failed: $malloc_line_csv does not exist. Check if the query ran correctly."
    exit 1
fi
if [ "$(grep -c . "$malloc_line_csv")" -ne 3 ]; then
    echo "External Verify Test for mallocgc time mapping Failed"
    echo "The content of mallocgc time csv file is:"
    cat "$malloc_line_csv"
    exit 1
fi

# check malloc_line's cumTime > 0
if awk -F',' 'NR>1 && $2+0 <= 0 { exit 1 }' "$malloc_line_csv"; then
    :
else
    echo "External Verify Test for mallocgc time mapping Failed: second column contains value <= 0"
    echo "The content of mallocgc time csv file is:"
    cat "$malloc_line_csv"
    exit 1
fi

echo "Done!"
