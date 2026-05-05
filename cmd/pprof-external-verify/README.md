# pprof-external-verify

A verification tool for validating data integrity when importing pprof profile data into CodeQL external predicates.

## Overview

There are bridge [external predicates](../../qlsrc/pprof_ext/profile_ext.qll) to import pprof exported data to CodeQL world. When encapsulating these external predicates, some data loss may occur. This verification tool is designed to detect and report such data loss by comparing counts between CodeQL query results and the original CSV files.

## Verification Types

| Verification Type | Query File                                                                              | Description                                                          |
| ----------------- | -------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| instance_count    | [instance_cnt_test.ql](../../qlsrc/pprof_ext/external_verify/instance_cnt_test.ql)     | Verifies count consistency of Sample, Location, Line, Function, etc. |
| relation_count    | TODO                                                                                   | Verifies count consistency of sample_to_location_id, sample_to_value, etc. |

## Usage

```bash
go run ./cmd/pprof-external-verify <codeql_database> <verification_type>
```

### Parameters

| Parameter              | Description                                                                 |
| ---------------------- | --------------------------------------------------------------------------- |
| `codeql_database`      | Path to the CodeQL database directory containing the pprof external data   |
| `verification_type`    | Type of verification to perform: `instance_count` or `relation_count`      |

### Example

```bash
# Verify instance counts in a CodeQL database
go run ./cmd/pprof-external-verify /path/to/codeql/database instance_count
```

## Verification Process

### instance_count

1. **Query Execution**: Runs the [instance_cnt_test.ql](../../qlsrc/pprof_ext/external_verify/instance_cnt_test.ql) query against the CodeQL database to obtain counts of:
   - `sample_cnt`: Number of Sample instances
   - `function_cnt`: Number of Function instances
   - `label_cnt`: Number of Label instances
   - `line_cnt`: Number of Line instances
   - `location_cnt`: Number of Location instances
   - `mapping_cnt`: Number of Mapping instances
   - `string_table_cnt`: Number of entries in the string table
   - `value_type_cnt`: Number of ValueType instances

2. **CSV Comparison**: Matches these counts against the corresponding CSV files in `<codeql_database>/ext/`:
   - Each CSV file (e.g., `sample.csv`, `location.csv`) contains raw data exported from pprof. The csv files don't contain header.
   - Counts are verified to ensure no data was lost during the external predicate encapsulation

3. **Output Format**: Results are output in CSV format with headers:
   ```
   "entity_type","codeql_count","csv_count","status"
   ```
   Where `status` is either `MATCH` or `MISMATCH`

## Exit Codes

| Exit Code | Description                                      |
| --------- | ------------------------------------------------ |
| 0         | All verifications passed successfully            |
| 1         | Verification failed (counts do not match)       |
| 2         | Invalid arguments or internal error              |
