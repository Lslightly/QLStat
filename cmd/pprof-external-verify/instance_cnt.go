package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Lslightly/qlstat/utils"
)

// entityTypes defines the list of entity types to be verified
var entityTypes = []string{
	"sample", "function", "label", "line",
	"location", "mapping", "string_table", "value_type",
}

// EntityCount represents the count of a specific entity type
type EntityCount struct {
	EntityType  string
	CodeQLCount int64
	CSVCount    int64
	Status      string
}

// verifyInstanceCount verifies that instance counts from CodeQL queries
// match the counts from the original CSV files in the external data directory.
// Returns true if all counts match, false otherwise.
func verifyInstanceCount(dbPath, extDir string) (bool, error) {
	// Step 1: Execute the CodeQL query to get instance counts
	codeQLCounts, err := executeInstanceCountQuery(dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to execute CodeQL query: %w", err)
	}

	// Step 2: Count rows in each CSV file
	csvCounts, err := countCSVEntries(extDir)
	if err != nil {
		return false, fmt.Errorf("failed to count CSV entries: %w", err)
	}

	// Step 3: Compare counts and generate verification results
	allMatch := compareAndPrintResults(codeQLCounts, csvCounts)

	return allMatch, nil
}

// externalPredicates defines all external predicates that need to be loaded
// from CSV files when executing the pprof-related queries.
// These predicates correspond to the pprof protobuf schema and are used
// to import pprof data into CodeQL's external predicate framework.
var externalPredicates = []string{
	"profile", "value_type", "sample", "sample_to_location_id", "sample_to_value",
	"sample_to_label", "label", "mapping", "location", "location_to_line",
	"line", "function", "string_table", "profile_to_sample_type",
	"profile_to_sample", "profile_to_mapping", "profile_to_location",
	"profile_to_function", "profile_to_string_table", "profile_to_comment",
}

// generateExternalOptions generates --external flags for CodeQL query execution.
// Each flag maps a predicate name to its corresponding CSV file path.
// This allows CodeQL to load external data from CSV files during query execution.
func generateExternalOptions(dbPath string) []string {
	extDir := filepath.Join(dbPath, "ext")
	args := make([]string, 0, len(externalPredicates))

	for _, predicate := range externalPredicates {
		csvPath := filepath.Join(extDir, predicate+".csv")
		args = append(args, "--external="+predicate+"="+csvPath)
	}

	return args
}

// executeInstanceCountQuery runs the instance_cnt_test.ql query against the CodeQL database
// and returns a map of entity type to count.
func executeInstanceCountQuery(dbPath string) (map[string]int64, error) {
	projRoot := utils.ProjectRoot()
	// Create a temporary file for the query output
	tmpFile, err := os.CreateTemp("", "instance_cnt_*.csv")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()
	bqrsFile := tmpFile.Name() + ".bqrs"

	// Generate external predicate options
	externalOpts := generateExternalOptions(dbPath)

	// Execute the CodeQL query and decode to CSV format
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // 5 minutes timeout
	defer cancel()

	// Build command arguments
	cmdArgs := []string{
		"query", "run",
		"-d=" + dbPath,
		"--search-path=" + filepath.Join(projRoot, "qlsrc/lib"),
		filepath.Join(projRoot, "qlsrc/pprof_ext/external_verify/instance_cnt_test.ql"),
		"--output=" + bqrsFile,
	}
	cmdArgs = append(cmdArgs, externalOpts...)

	cmd := exec.CommandContext(ctx, "codeql", cmdArgs...)

	cmd.Dir = projRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("codeql query run failed: %w", err)
	}
	defer os.Remove(bqrsFile)

	// Decode the bqrs result to CSV
	cmd = exec.CommandContext(ctx, "codeql",
		"bqrs", "decode",
		"--format=csv",
		bqrsFile,
		"--output="+tmpFile.Name(),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("codeql bqrs decode failed: %w", err)
	}

	// Parse the CSV output
	return parseCodeQLCSV(tmpFile.Name())
}

// parseCodeQLCSV reads the CSV output from CodeQL and extracts entity counts.
// The CSV header is: "sample_cnt","function_cnt","label_cnt","line_cnt","location_cnt","mapping_cnt","string_table_cnt","value_type_cnt"
func parseCodeQLCSV(csvPath string) (map[string]int64, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 1 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	// First row is header, second row contains the counts
	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file has no data rows")
	}

	header := records[0]
	data := records[1]

	// Generate expected columns by appending "_cnt" suffix to each entity type
	expectedColumns := make([]string, len(entityTypes))
	for i, entityType := range entityTypes {
		expectedColumns[i] = entityType + "_cnt"
	}

	if len(header) != len(expectedColumns) {
		return nil, fmt.Errorf("unexpected CSV header: expected %d columns, got %d", len(expectedColumns), len(header))
	}

	counts := make(map[string]int64)

	for i, colName := range header {
		if i >= len(data) {
			return nil, fmt.Errorf("data row shorter than header at column %d", i)
		}

		// Remove surrounding quotes from column name (CSV format)
		colName = strings.Trim(colName, "\"")

		// Assert that column name ends with "_cnt" suffix
		if !strings.HasSuffix(colName, "_cnt") {
			return nil, fmt.Errorf("column name '%s' does not have '_cnt' suffix", colName)
		}

		// Derive entity type by removing the "_cnt" suffix
		entityType := strings.TrimSuffix(colName, "_cnt")

		count, err := strconv.ParseInt(strings.Trim(data[i], "\""), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse count for %s: %w", colName, err)
		}

		// Store counts with both the column name and the derived entity type
		counts[colName] = count
		counts[entityType] = count
	}

	return counts, nil
}

// countCSVEntries counts the number of rows in each relevant CSV file
// in the external data directory.
func countCSVEntries(extDir string) (map[string]int64, error) {
	counts := make(map[string]int64)

	for _, entityType := range entityTypes {
		csvFileName := entityType + ".csv"
		csvPath := filepath.Join(extDir, csvFileName)

		count, err := countCSVLines(csvPath)
		if err != nil {
			// If file doesn't exist, treat count as 0
			if os.IsNotExist(err) {
				counts[entityType] = 0
				continue
			}
			return nil, fmt.Errorf("failed to count %s.csv: %w", entityType, err)
		}

		counts[entityType] = count
	}

	return counts, nil
}

// countCSVLines counts the number of lines (rows) in a CSV file.
// The file does not contain a header row.
func countCSVLines(csvPath string) (int64, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var count int64 = 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines
		if line != "" {
			count++
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading file: %w", err)
	}

	return count, nil
}

// compareAndPrintResults compares CodeQL counts with CSV counts and prints results.
// Returns true if all counts match.
func compareAndPrintResults(codeQLCounts, csvCounts map[string]int64) bool {
	// Print header
	fmt.Printf("\n%-15s %-15s %-15s %-10s\n", "ENTITY_TYPE", "CODEQL_COUNT", "CSV_COUNT", "STATUS")
	fmt.Println(strings.Repeat("-", 60))

	allMatch := true

	for _, entityType := range entityTypes {
		codeQLCount := codeQLCounts[entityType+"_cnt"]
		csvCount := csvCounts[entityType]

		status := "MATCH"
		if codeQLCount != csvCount {
			status = "MISMATCH"
			allMatch = false
		}

		fmt.Printf("%-15s %-15d %-15d %-10s\n", entityType, codeQLCount, csvCount, status)
	}

	fmt.Println(strings.Repeat("-", 60))

	return allMatch
}
