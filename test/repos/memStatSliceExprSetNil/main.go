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
