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

func main() {
	good()
	bad()
	goodLog()
	badLog()
}

type S struct{}
type sliceS []*S

type ptr *int
type slice []ptr

type Log struct {
	buffer slice
}

func good() {
	var s slice
	for i := 0; i < 8; i++ {
		s = append(s, &i)
	}
	s[0] = nil
	s = s[1:]
	for _, elem := range s {
		print(*elem)
	}
}

func bad() {
	var s slice
	for i := 0; i < 8; i++ {
		s = append(s, &i)
	}
	s = s[1:]
	for _, elem := range s {
		print(*elem)
	}
}

func goodLog() {
	var log Log
	log.buffer = make(slice, 2)
	log.buffer[0] = nil
	log.buffer = log.buffer[1:]
}

func badLog() {
	var log Log
	log.buffer = make(slice, 2)
	log.buffer = log.buffer[1:]
}

func cutBack() {
	var s slice
	for i := 0; i < 8; i++ {
		s = append(s, &i)
	}
	s[0] = nil
	s = s[:2]
	for _, elem := range s {
		print(*elem)
	}
}

func goodSliceS() {
	var s sliceS
	for i := 0; i < 8; i++ {
		s = append(s, new(S))
	}
	s[0] = nil
	s = s[1:]
	for _, elem := range s {
		print(*elem)
	}
}

func badSliceS() {
	var s sliceS
	for i := 0; i < 8; i++ {
		s = append(s, new(S))
	}
	s = s[1:]
	for _, elem := range s {
		print(*elem)
	}
}
