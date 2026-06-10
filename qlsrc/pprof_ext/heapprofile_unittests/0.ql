import pprof_ext.heapprofile

from HeapProfile hp
select hp.allocObjectsSum().toString() as alloc_objects
