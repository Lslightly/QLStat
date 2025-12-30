package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lslightly/qlstat/utils"
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
	utils.MkdirAll(OutDir)
	if Opts.movedToHeap {
		func() {
			outfile := utils.CreateFile(filepath.Join(OutDir, "movedToHeap.csv"))
			defer outfile.Close()
			fmt.Fprint(outfile, strings.Join(movedToHeapHandle(lines), "\n"))
		}()
	}
}
