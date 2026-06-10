import pprof_ext.heapprofile

from HeapProfile hp
select bytesInUnit(hp.allocSpaceOfLine("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8", 20), "KB", 4) as malloc8_line37_alloc_bytes
