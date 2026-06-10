import pprof_ext.heapprofile

from HeapProfile hp
select hp.inuseObjectsPercent("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8", 2) as malloc8_inuse_obj_pct
