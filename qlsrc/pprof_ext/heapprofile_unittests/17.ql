import pprof_ext.heapprofile

from HeapProfile hp
select bytesInUnit(hp.inuseSpaceOfLine("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8", 20), "KB", 4) as malloc8_line37_inuse_bytes
