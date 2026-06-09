import heapprofile

from HeapProfile hp
select hp.allocObjectsSum().toInt() as alloc_objects,
    bytesInUnit(hp.allocSpaceSum(), "KB", 4) as alloc_bytes,
    hp.inuseObjectsSum().toInt() as inuse_objects,
    bytesInUnit(hp.inuseObjectsSum(), "KB", 4) as inuse_space,
    hp.allocObjectsOfFunc("malloctest.BenchmarkMalloc8").toInt() as malloc8_alloc_objects,
    bytesInUnit(hp.allocSpaceOfFunc("malloctest.BenchmarkMalloc8"), "KB", 4) as malloc8_alloc_bytes,
    hp.inuseObjectsOfFunc("malloctest.BenchmarkMalloc8").toInt() as malloc8_inuse_objects,
    bytesInUnit(hp.inuseSpaceOfFunc("malloctest.BenchmarkMalloc8"), "KB", 4) as malloc8_inuse_bytes,
    // flat (self) values — only samples where the function is at the top of the stack
    hp.allocObjectsFlatOfFunc("malloctest.BenchmarkMalloc8").toInt() as malloc8_flat_alloc_objects,
    bytesInUnit(hp.allocSpaceFlatOfFunc("malloctest.BenchmarkMalloc8"), "KB", 4) as malloc8_flat_alloc_bytes,
    hp.inuseObjectsFlatOfFunc("malloctest.BenchmarkMalloc8").toInt() as malloc8_flat_inuse_objects,
    bytesInUnit(hp.inuseSpaceFlatOfFunc("malloctest.BenchmarkMalloc8"), "KB", 4) as malloc8_flat_inuse_bytes,
    // percentages
    hp.allocObjectsPercent("malloctest.BenchmarkMalloc8", 2) as malloc8_alloc_obj_pct,
    hp.allocSpacePercent("malloctest.BenchmarkMalloc8", 2) as malloc8_alloc_bytes_pct,
    hp.inuseObjectsPercent("malloctest.BenchmarkMalloc8", 2) as malloc8_inuse_obj_pct,
    hp.inuseSpacePercent("malloctest.BenchmarkMalloc8", 2) as malloc8_inuse_bytes_pct,
    // line-level: cumulative allocations at specific lines
    bytesInUnit(hp.allocSpaceOfLine("malloctest.BenchmarkMalloc8", 20), "KB", 4) as malloc8_line37_alloc_bytes,
    bytesInUnit(hp.inuseSpaceOfLine("malloctest.BenchmarkMalloc8", 20), "KB", 4) as malloc8_line37_inuse_bytes,
    // block size
    hp.blockSizeSumOfFunc("malloctest.BenchmarkMalloc8").toInt() as malloc8_block_size
