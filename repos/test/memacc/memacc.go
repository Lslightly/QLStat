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

type S struct {
	f int
}

type WrapS struct {
	g  int
	s1 S
}

type WrapWrapS struct {
	h  int
	s2 WrapS
}

func (s S) Print(pa *int, b, c int) {
	println(s.f + *pa + b + c)
}

func (wws *WrapWrapS) Print() {
	wws.s2.s1.f = 10
	println(wws.s2.s1.f)
}

func main() {
	var a int
	var b int = 2
	var arr [3]int = [3]int{10, 20, 30}
	var slice []int = []int{40, 50}
	c := 3
	pa := &a
	ppa := &pa
	c = 2
	**ppa = 4
	*pa = 1
	s := &S{f: 20}
	s.Print(pa, b, c)
	f := func(a int) {
		println(a)
	}
	f(*pa)
	println(arr[1])
	println(slice[0] == 2*arr[1])
	*pa = 1
	d := 1
	a, b = c, d
}
