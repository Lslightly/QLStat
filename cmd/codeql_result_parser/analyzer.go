package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/Lslightly/qlstat/utils"
	"github.com/schollz/progressbar/v3"
)

type Analyzer interface {
	// baseAnalyzer common part
	SetWorkDir(workDir string)
	Dump()
	outfilePath() string

	Analyze(csvs []CodeQLCSV)
	String() string
	name() string
}

var cleanedDir map[string]bool

func init() {
	cleanedDir = make(map[string]bool)
}

// 基础analyzer，其他analyzer使用其SetWorkDir,
type IOHandler struct {
	workDir string
	outDir  string
	header  []string
	child   Analyzer
}

func (this *IOHandler) SetWorkDir(workDir string) {
	this.workDir = workDir
	this.outDir = filepath.Join(workDir, "analyze")
	if _, exists := cleanedDir[this.outDir]; !exists {
		fmt.Println("recreate out dir", this.outDir)
		utils.Remkdir(this.outDir)
	}
	cleanedDir[this.outDir] = true
}

func (this *IOHandler) Dump() {
	fmt.Println("dumping", this.outfilePath())
	ofile, err := os.Create(this.outfilePath())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprint(ofile, this.child)

	ofile.Close()
}

func (this *IOHandler) outfilePath() string {
	return filepath.Join(this.outDir, this.child.name()+".csv")
}

func (this *IOHandler) getAndTrimHeader(csvs []CodeQLCSV) (dataWithoutHeader []CodeQLCSV) {
	if len(csvs) == 0 {
		log.Fatal("no csv passed in getHeader")
	}
	this.header = csvs[0].data[0]

	bar := progressbar.Default(int64(len(csvs)), "get and trim header")
	defer bar.Close()

	dataWithoutHeader = make([]CodeQLCSV, len(csvs))
	for i, csv := range csvs {
		if !sameHeader(this.header, csv.data[0]) {
			log.Fatalln("The header of", csvs[0].path, "is different from", csv.path)
		}
		dataWithoutHeader[i] = csv
		dataWithoutHeader[i].data = dataWithoutHeader[i].data[1:] // trim header
		bar.Add(1)
	}
	return
}

func sameHeader(this []string, that []string) bool {
	if len(this) != len(that) {
		return false
	}
	for i := 0; i < len(this); i++ {
		if this[i] != that[i] {
			return false
		}
	}
	return true
}

func argConstraint(argNum int, name string, args ...interface{}) error {
	if len(args) != argNum {
		return errors.New(fmt.Sprintln(name, "need", argNum, "arguments but", len(args), "is provided", args))
	}
	return nil
}

// count all entries in csvs
type Counter struct {
	IOHandler
	cnt int
}

func newCounter(args ...interface{}) (*Counter, error) {
	if err := argConstraint(0, "Counter", args...); err != nil {
		return nil, err
	}
	this := &Counter{
		cnt: 0,
	}
	this.child = this
	return this, nil
}

func (this *Counter) Analyze(csvs []CodeQLCSV) {
	csvs = this.getAndTrimHeader(csvs)

	bar := progressbar.Default(int64(len(csvs)))
	defer bar.Close()
	for _, csv := range csvs {
		this.cnt += len(csv.data)
		bar.Add(1)
	}
}

func (this *Counter) String() string {
	res := fmt.Sprintln("cnt")
	res += fmt.Sprintln(strconv.Itoa(this.cnt))
	return res
}

func (this *Counter) name() string {
	return "count"
}

// group by `groupColumn` then count
type GroupByCounter struct {
	IOHandler
	groupColumn uint64
	value2cnt   map[interface{}]int
}

func newGroupByCounter(args ...interface{}) (*GroupByCounter, error) {
	if err := argConstraint(1, "GroupByCounter", args...); err != nil {
		return nil, err
	}
	if groupColumn, ok := args[0].(uint64); ok {
		this := &GroupByCounter{
			groupColumn: groupColumn,
			value2cnt:   make(map[interface{}]int),
		}
		this.child = this
		return this, nil
	} else {
		return nil, errors.New(fmt.Sprintln("GroupByCounter need uint64 argument but", reflect.TypeOf(args[0]), "is provided"))
	}
}

func (this *GroupByCounter) valueAddOne(v string) {
	if _, ok := this.value2cnt[v]; !ok {
		this.value2cnt[v] = 1
	} else {
		this.value2cnt[v]++
	}
}

func (this *GroupByCounter) Analyze(csvs []CodeQLCSV) {
	csvs = this.getAndTrimHeader(csvs)

	bar := progressbar.Default(int64(len(csvs)))
	defer bar.Close()
	for _, csv := range csvs {
		for _, row := range csv.data {
			this.valueAddOne(row[this.groupColumn])
		}
		bar.Add(1)
	}
}

func (this *GroupByCounter) String() string {
	res := fmt.Sprintf("%s,%s\n", this.header[this.groupColumn], "cnt")
	for v, cnt := range this.value2cnt {
		res += fmt.Sprintf("%s,%d\n", v, cnt)
	}
	return res
}

func (this *GroupByCounter) name() string {
	return fmt.Sprintf("groupBy%dCnt", this.groupColumn)
}

// concat all csvs to one csv and append a column of repo name
type Concator struct {
	IOHandler
	data [][]string
}

func newConcator(args ...interface{}) (*Concator, error) {
	if err := argConstraint(0, "Concator", args...); err != nil {
		return nil, err
	}
	this := &Concator{}
	this.child = this
	return this, nil
}

func (this *Concator) Analyze(csvs []CodeQLCSV) {
	csvs = this.getAndTrimHeader(csvs)
	this.header = append(this.header, `"repoName"`)
	this.data = make([][]string, 0, len(csvs))
	for _, csv := range csvs {
		for rowIdx := range csv.data { // add repoName at the end column
			csv.data[rowIdx] = append(csv.data[rowIdx], csv.GetRepoName())
		}
		this.data = append(this.data, csv.data...)
	}
}

func (this *Concator) String() string {
	res := strings.Join(this.header, ",") + "\n"
	for _, lineElems := range this.data {
		res += fmt.Sprintln(strings.Join(lineElems, ","))
	}
	return res
}

func (this *Concator) name() string {
	return "concat"
}
