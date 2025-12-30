package main

/*
escape analysis adapter

The adapter will adapt the escape analysis -m=2 output to CSV files
and store these files in $DBRoot/<path/to/repo>/ext/<pred>.csv. Then
these adapted results can be used by `codeql query run --external`
to extend the ability of CodeQL.

The input of the adapter is:
1. the predicate to generate
2. the output csv dir
3. the input escape analysis log

The format of output csv file should follow ext_preds.qll in
qlsrc/escape_ext/ext_preds.qll.
*/
