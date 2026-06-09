import pprof_ext.profile_ext

from int sample_id, int index, QlBuiltins::BigInt value
where sample_to_value(sample_id, index, value)
select sample_id, index, value
