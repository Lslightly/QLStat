package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Lslightly/qlstat/utils"
)

var (
	OutDir  string
	SrcRoot string
	Opts    Options
)

type Options struct {
	movedToHeap bool
}

func init() {
	flag.StringVar(&OutDir, "dir", "./", "(required) directory to store output csv")
	flag.StringVar(&SrcRoot, "src", "./", "(required) source root directory. Paths will be adjusted to absolute path according to source root directory")
	flag.BoolVar(&Opts.movedToHeap, "movedToHeap", false, "enable \"moved to heap\" dump")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Escape Analysis Adapter\nUsage: go run cmd/escape_adapter [options...] <escape analysis log>")
		flag.PrintDefaults()
	}
}

type LineConsumer func(i int, line string) bool
type LineGenerator func(yield LineConsumer)

func createLineGen(logpath string) (res LineGenerator) {
	return func(yield LineConsumer) {
		f := utils.OpenFile(logpath)
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for i := 0; scanner.Scan(); i++ {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if !yield(i, line) {
				break
			}
		}
	}
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	logpath := flag.Arg(0)
	utils.MkdirAll(OutDir)
	if Opts.movedToHeap {
		func() {
			outfile := utils.CreateFile(filepath.Join(OutDir, "movedToHeap.csv"))
			defer outfile.Close()
			fmt.Fprint(outfile, strings.Join(movedToHeapHandle(createLineGen(logpath)), "\n"))
		}()
	}
}
