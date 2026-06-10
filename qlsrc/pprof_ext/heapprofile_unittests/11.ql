import pprof_ext.heapprofile

from HeapProfile hp
select bytesInUnit(hp.inuseSpaceFlatOfFunc("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8"), "KB", 4) as malloc8_flat_inuse_bytes
