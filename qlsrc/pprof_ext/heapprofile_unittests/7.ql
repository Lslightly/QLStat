import pprof_ext.heapprofile

from HeapProfile hp
select bytesInUnit(hp.inuseSpaceOfFunc("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8"), "KB", 4) as malloc8_inuse_bytes
