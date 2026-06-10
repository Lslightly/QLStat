import pprof_ext.heapprofile

from HeapProfile hp
select bytesInUnit(hp.allocSpaceSum(), "KB", 4) as alloc_bytes
