package main

import "testing"

var alwaysFalse bool
var escapeSink any

func Escape[T any](x T) T {
	if alwaysFalse {
		escapeSink = x
	}
	return x
}

func BenchmarkNewS(b *testing.B) {
	for b.Loop() {
		ps := newS()
		Escape(ps)
	}
}
