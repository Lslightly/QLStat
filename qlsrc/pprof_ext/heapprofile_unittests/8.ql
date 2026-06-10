import pprof_ext.heapprofile

from HeapProfile hp
select hp.allocObjectsFlatOfFunc("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8").toString() as malloc8_flat_alloc_objects
