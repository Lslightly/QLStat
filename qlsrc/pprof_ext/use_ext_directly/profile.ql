import pprof_ext.profile_ext

from int id, int drop_frames, int keep_frames, QlBuiltins::BigInt time_nanos, QlBuiltins::BigInt duration_nanos, int period_type, QlBuiltins::BigInt period, int default_sample_type, int doc_url
where profile(id, drop_frames, keep_frames, time_nanos, duration_nanos, period_type, period, default_sample_type, doc_url)
select id, drop_frames, keep_frames, time_nanos.toString(), duration_nanos.toString(), period_type, period.toString(), default_sample_type, doc_url
