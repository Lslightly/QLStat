import pprof_ext.heapprofile

from HeapProfile hp
select hp.inuseSpacePercent("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8", 2) as malloc8_inuse_bytes_pct
