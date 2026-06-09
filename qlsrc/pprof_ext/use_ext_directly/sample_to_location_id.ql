import pprof_ext.profile_ext

from int sample_id, int index, int location_id
where sample_to_location_id(sample_id, index, location_id)
select sample_id, index, location_id
