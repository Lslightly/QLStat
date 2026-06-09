import pprof_ext.profile_ext

from int location_id, int index, int line_id
where location_to_line(location_id, index, line_id)
select location_id, index, line_id
