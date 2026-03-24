package malloctest

import (
	"testing"
	"unsafe"
)

var alwaysFalse bool
var escapeSink any

func Escape[T any](x T) T {
	if alwaysFalse {
		escapeSink = x
	}
	return x
}

func BenchmarkMalloc8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := new(int64)
		Escape(p)
	}
}

func BenchmarkMalloc16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := new([2]int64)
		Escape(p)
	}
}

func BenchmarkMallocTypeInfo8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := new(struct {
			p [8 / unsafe.Sizeof(uintptr(0))]*int
		})
		Escape(p)
	}
}

func BenchmarkMallocTypeInfo16(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := new(struct {
			p [16 / unsafe.Sizeof(uintptr(0))]*int
		})
		Escape(p)
	}
}

type LargeStruct struct {
	x [16][]byte
}

var largeStructRunCnt int = 0

func BenchmarkMallocLargeStruct(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := make([]LargeStruct, 2)
		Escape(p)
	}
	b.StopTimer()
}
