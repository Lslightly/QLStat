import pprof_ext.profile_ext

from int sample_id, int index, int label_id
where sample_to_label(sample_id, index, label_id)
select sample_id, index, label_id
