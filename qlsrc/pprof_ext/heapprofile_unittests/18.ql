import pprof_ext.heapprofile

from HeapProfile hp
select hp.blockSizeSumOfFunc("github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc8").toString() as malloc8_block_size
