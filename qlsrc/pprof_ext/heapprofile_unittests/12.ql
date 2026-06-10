import pprof_ext.heapprofile

from HeapProfile hp
select hp.allocObjectsPercent("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8", 2) as malloc8_alloc_obj_pct
