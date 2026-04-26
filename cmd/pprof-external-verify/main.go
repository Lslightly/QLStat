package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Verification types supported by the tool
const (
	VerificationTypeInstanceCount = "instance_count"
	VerificationTypeRelationCount = "relation_count"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "pprof-external-verify: Verify data integrity of pprof external predicates in CodeQL databases")
		fmt.Fprintln(os.Stderr, "\nUsage: pprof-external-verify [options] <codeql_database> <verification_type>")
		fmt.Fprintln(os.Stderr, "\nVerification types:")
		fmt.Fprintln(os.Stderr, "  instance_count   - Verify count consistency of Sample, Location, Line, Function, etc.")
		fmt.Fprintln(os.Stderr, "  relation_count   - Verify count consistency of sample_to_location_id, sample_to_value, etc. (TODO)")
		fmt.Fprintln(os.Stderr, "\nOptions:")
		flag.PrintDefaults()
	}

	// Parse command line arguments
	flag.Parse()

	// Validate argument count
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	// Extract arguments
	dbPath := flag.Arg(0)
	verificationType := flag.Arg(1)

	// Validate verification type
	if verificationType != VerificationTypeInstanceCount && verificationType != VerificationTypeRelationCount {
		fmt.Fprintf(os.Stderr, "Error: Unknown verification type '%s'\n", verificationType)
		flag.Usage()
		os.Exit(2)
	}

	externalDir := filepath.Join(dbPath, "ext")

	// Execute the appropriate verification based on type
	var success bool
	var err error

	switch verificationType {
	case VerificationTypeInstanceCount:
		success, err = verifyInstanceCount(dbPath, externalDir)
	case VerificationTypeRelationCount:
		fmt.Fprintln(os.Stderr, "Error: relation_count verification is not yet implemented")
		os.Exit(2)
	}

	// Handle errors
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error during verification: %v\n", err)
		os.Exit(2)
	}

	// Set exit code based on verification result
	if !success {
		fmt.Println("\nVerification FAILED: Count mismatches detected")
		os.Exit(1)
	}

	fmt.Println("\nVerification PASSED: All counts match")
	os.Exit(0)
}
