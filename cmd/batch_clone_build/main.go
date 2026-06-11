package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Lslightly/qlstat/config"
)

type Options struct {
	disableClone       bool
	disableBuild       bool
	disableExternalGen bool
}

var opt Options

func init() {
	flag.BoolVar(&opt.disableClone, "noclone", false, "disable clone")
	flag.BoolVar(&opt.disableBuild, "nobuild", false, "disable build")
	flag.BoolVar(&opt.disableExternalGen, "noextgen", false, "disable generating database for external predicates")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/batch_clone_build <yaml file>")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	cfg := config.UnmarshalArtifact(flag.Arg(0))
	dirSetup(cfg)
	if !opt.disableClone {
		batchClone(cfg)
	}
	if !opt.disableBuild {
		batchBuild(cfg)
	}
	if !opt.disableExternalGen {
		batchExternalGen(cfg)
	}
}
