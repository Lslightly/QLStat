import pprof_ext.heapprofile

from HeapProfile hp
select hp.inuseObjectsOfFunc("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8").toString() as malloc8_inuse_objects
