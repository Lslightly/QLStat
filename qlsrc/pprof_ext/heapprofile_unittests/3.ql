import pprof_ext.heapprofile

from HeapProfile hp
select bytesInUnit(hp.inuseObjectsSum(), "KB", 4) as inuse_space
