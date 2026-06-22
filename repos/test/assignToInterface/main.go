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

import "strconv"

type Iface interface {
	String() string
}

type Eface interface{}

type S struct {
	a int
}

func (s S) String() string {
	return strconv.Itoa(s.a)
}

func main() {
	assignToEface()
	assignToIface()
}

func assignToEface() {
	var eface Eface
	var i2 interface{}
	var iface Iface
	eface = iface
	eface = true
	a := 1
	eface = a
	str := "abcdef"
	eface = str
	arr := [2]int{1, 2}
	eface = arr
	slice := []int{1, 2, 3}
	eface = slice
	s := S{
		a: 2,
	}
	eface = s
	ptr := &a
	eface = ptr
	f := func() {
		println(a)
		a = 2
		str = "fdas"
		println(str)
	}
	eface = f
	i2 = s
	eface = i2
	m := map[int]string{
		1: "1",
		2: "2",
	}
	eface = m
	ch := make(chan bool, 2)
	eface = ch
	println(eface)
}

func assignToIface() {
	var iface Iface
	s := S{
		a: 2,
	}
	iface = s
	println(iface)
}
