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
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/goccy/go-yaml"
	"github.com/schollz/progressbar/v3"
)

// Flags
var (
	configPath string
	resultRoot string
)

func init() {
	flag.StringVar(&configPath, "c", "qlSumConfig.yaml", "the analyzer yaml configuration file")
	flag.Usage = func() {
		fmt.Println("NOTICE: Please use `-collect` option of codeql_qdriver and then use database to import the collected csv. Most functions here can be done by SQL in mature database.\naccept the root directory path of codeql result as the last argument")
		flag.PrintDefaults()
	}
}

// ql name to csv
type QL2CSVsTy map[string][]CodeQLCSV

// ql name to analyzer
var ql2analyzer map[string][]Analyzer = make(map[string][]Analyzer)

func main() {
	checkCmd()
	parseAnalyzerConfig()
	ql2csvs := getQL2CSVsMap()
	for qlName, csvs := range ql2csvs {
		fmt.Println(qlName)
		analyzers := ql2analyzer[qlName]
		for _, analyzer := range analyzers {
			analyzer.SetWorkDir(path.Join(resultRoot, qlName))
			analyzer.Analyze(csvs)
			analyzer.Dump()
		}
	}
}

func parseAnalyzerConfig() {
	bs, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalln(err)
	}

	var qlconfig ConfigTy
	err = yaml.Unmarshal(bs, &qlconfig)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range qlconfig.Entries {
		qlname, analyzer := entry.resolve()
		fmt.Println("get", analyzer.name(), "accumulator for", qlname)
		if _, exists := ql2analyzer[qlname]; !exists {
			ql2analyzer[qlname] = make([]Analyzer, 0)
		}
		ql2analyzer[qlname] = append(ql2analyzer[qlname], analyzer)
	}
}

func checkCmd() {
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}
	resultRoot = flag.Arg(0)
	if _, err := os.Stat(resultRoot); errors.Is(err, os.ErrExist) {
		log.Fatalln(resultRoot, "does not exist.")
		os.Exit(1)
	}
}

/*
return a map from ql root path to csvs paths
*/
func getQL2CSVsMap() (ql2csvs QL2CSVsTy) {
	ql2csvs = make(QL2CSVsTy)
	for qlName := range ql2analyzer {
		fmt.Println("getting ql2csvs for", qlName)
		qlDir := filepath.Join(resultRoot, qlName)
		entries, err := os.ReadDir(qlDir)
		if err != nil {
			fmt.Println(err)
			continue
		}
		csvs := make([]CodeQLCSV, 0, len(entries)/2)

		bar := progressbar.Default(int64(len(entries)))

		for _, entry := range entries {
			bar.Add(1)
			if !(!entry.IsDir() && filepath.Ext(entry.Name()) == ".csv") {
				continue
			}
			csv, err := NewCSV(filepath.Join(qlDir, entry.Name()), qlName)
			if err != nil {
				continue
			}
			csvs = append(csvs, csv)
		}
		ql2csvs[qlName] = csvs
		fmt.Println("get", len(csvs), "csv files for", qlName)

		bar.Close()
	}
	return
}
