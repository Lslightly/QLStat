package pkgcall_test

import "testing"

func BenchmarkA(t *testing.B) {
	t.StartTimer()
	defer t.StopTimer()
	print("a")
}
