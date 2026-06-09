import pprof_ext.profile_ext

from int id, QlBuiltins::BigInt memory_start, QlBuiltins::BigInt memory_limit, QlBuiltins::BigInt file_offset, int filename, int build_id, boolean has_functions, boolean has_filenames, boolean has_line_numbers, boolean has_inline_frames
where mapping(id, memory_start, memory_limit, file_offset, filename, build_id, has_functions, has_filenames, has_line_numbers, has_inline_frames)
select id, memory_start, memory_limit, file_offset, filename, build_id, has_functions, has_filenames, has_line_numbers, has_inline_frames
