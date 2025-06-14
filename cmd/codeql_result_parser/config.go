package main

import (
	"log"
)

type ConfigTy struct {
	Entries []EntryTy `yaml:"entries"`
}

type EntryTy struct {
	Name string      `yaml:"name"`
	Cnt  CounterFnTy `yaml:"cnt"`
}

type CounterFnTy struct {
	Fn   string        `yaml:"fn"`
	Args []interface{} `yaml:"args"`
}

func (this EntryTy) resolve() (qlname string, analyzer Analyzer) {
	qlname = this.Name
	var err error
	switch this.Cnt.Fn {
	case "GroupByCounter":
		analyzer, err = newGroupByCounter(this.Cnt.Args...)
	case "Counter":
		analyzer, err = newCounter(this.Cnt.Args...)
	case "Concator":
		analyzer, err = newConcator(this.Cnt.Args...)
	default:
		log.Fatalln("Oops. unknown", this.Cnt.Fn)
	}
	if err != nil {
		log.Fatalln("errors occurs when resolving", qlname, ":", err)
	}
	return
}
