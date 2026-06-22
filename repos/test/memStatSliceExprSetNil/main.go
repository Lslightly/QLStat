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
	"fmt"
	"runtime"
	"time"
)

type LargeT struct {
	arr [100000]int
}

func main() {
	s := f()
	var m runtime.MemStats
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		runtime.ReadMemStats(&m)
		fmt.Println(m.Alloc)
		runtime.GC()
	}
	for i, elem := range s {
		elem.arr[0] = i
	}
}

func f() []*LargeT {
	len := 1000
	s := make([]*LargeT, len)
	for i := 0; i < len; i++ {
		ptr := new(LargeT)
		ptr.arr[0] = i
		s[i] = ptr
	}

	start := len - 200

	// comment the for loop for comparison
	for i := 0; i < start; i++ {
		s[i] = nil
	}
	s = s[start:]
	return s
}
