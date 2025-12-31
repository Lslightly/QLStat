package main

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func convint(s string) int {
	res, err := strconv.Atoi(s)
	if err != nil {
		log.Panicf("error converting %s to int: %v", s, err)
	}
	return res
}

func cleanpath(path string) string {
	absSrcRoot, err := filepath.Abs(SrcRoot)
	if err != nil {
		log.Panicf("error when converting SrcRoot: %s: %v", SrcRoot, err)
	}
	path = filepath.Clean(path)
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Clean(filepath.Join(absSrcRoot, path))
}

// path, startLine, startCol
func movedToHeapHandle(lineGen LineGenerator) (csvRows []string) {
	const pat string = "%s,%d,%d"
	for i, line := range lineGen {
		if !strings.Contains(line, "moved to heap") {
			continue
		}
		regex := regexp.MustCompile(`(.*?):(\d+):(\d+):`)
		matches := regex.FindStringSubmatch(line)
		if len(matches) == 0 {
			log.Printf("line %d with moved to heap but no match\n", i+1)
			continue
		}
		path, startLineStr, startColStr := matches[1], matches[2], matches[3]
		csvRows = append(csvRows, fmt.Sprintf(pat, cleanpath(path), convint(startLineStr), convint(startColStr)))
	}
	return
}
