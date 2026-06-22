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
