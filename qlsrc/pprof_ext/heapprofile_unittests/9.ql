import pprof_ext.heapprofile

from HeapProfile hp
select bytesInUnit(hp.allocSpaceFlatOfFunc("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8"), "KB", 4) as malloc8_flat_alloc_bytes
