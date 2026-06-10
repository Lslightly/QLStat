import pprof_ext.heapprofile

from HeapProfile hp
select hp.inuseObjectsSum().toString() as inuse_objects
