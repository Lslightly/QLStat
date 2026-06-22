// Copyright 2026 Qingwei Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Lslightly/qlstat/cmd/pprof2qlcsv/convert"
	"github.com/Lslightly/qlstat/utils"
	"github.com/google/pprof/profile"
)

var (
	outDir string
)

func init() {
	flag.StringVar(&outDir, "dir", "./", "directory to store output csv files")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "pprof2qlcsv: convert pprof profile to CodeQL external predicate CSV files\nUsage: pprof2qlcsv [options...] <profile>")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	profilePath := flag.Arg(0)
	p, err := parseProfile(profilePath)
	if err != nil {
		log.Fatalf("error parsing profile %s: %v", profilePath, err)
	}

	utils.MkdirAll(outDir)

	data := convert.Convert(p)
	if err := data.DumpCSV(outDir); err != nil {
		log.Fatalf("error writing csv: %v", err)
	}

	fmt.Printf("successfully converted %s to %s\n", profilePath, outDir)
}

// parseProfile 读取并解析 pprof 文件（支持 gzip 压缩的 protobuf 格式）。
func parseProfile(path string) (*profile.Profile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening profile: %w", err)
	}
	defer f.Close()

	p, err := profile.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing profile: %w", err)
	}
	return p, nil
}
