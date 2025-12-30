package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	OutDir string
	Opts   Options
)

type Options struct {
	movedToHeap bool
}

func init() {
	flag.StringVar(&OutDir, "dir", "./", "directory to store output csv")
	flag.BoolVar(&Opts.movedToHeap, "movedToHeap", false, "enable \"moved to heap\" dump")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Escape Analysis Adapter\nUsage: go run cmd/escape_adapter [options...] <escape analysis log>")
		flag.PrintDefaults()
	}
}

func readLinesFromFile(path string) []string {
	buf, err := os.ReadFile(path)
	if err != nil {
		log.Panicf("error when reading file %s: %v", path, err)
	}
	lines := strings.Split(string(buf), "\n")
	return lines
}

type HandleFunc func(lines []string)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	logpath := flag.Arg(0)
	lines := readLinesFromFile(logpath)
	if err := os.MkdirAll(OutDir, 0755); err != nil {
		log.Panicf("error occurs when creating dir %s: %v", OutDir, err)
	}
	if Opts.movedToHeap {
		func() {
			outfile, err := os.Create(filepath.Join(OutDir, "movedToHeap.csv"))
			if err != nil {
				log.Panicf("error occurs when creating file %s: %v", outfile.Name(), err)
			}
			defer outfile.Close()
			fmt.Fprint(outfile, strings.Join(movedToHeapHandle(lines), "\n"))
		}()
	}
}
