// Profile
external predicate profile_to_sample_type(int index, int sample_type_id);
external predicate profile_to_sample(int index, int sample_id);
external predicate profile_to_mapping(int index, int mapping_id);
external predicate profile_to_location(int index, int location_id);
external predicate profile_to_function(int index, int function_id);
external predicate profile_to_string_table(int index, int string_table_id);
external predicate profile_to_comment(int index, int comment_id);
external predicate profile(int id, int drop_frames, int keep_frames, int time_nanos, int duration_nanos, int period_type, int period, int default_sample_type, int doc_url);

// ValueType
external predicate value_type(int id, int type, int unit);

// Sample
external predicate sample_to_location_id(int sample_id, int index, int location_id);
external predicate sample_to_value(int sample_id, int index, int value);
external predicate sample_to_label(int sample_id, int index, int label_id);

// Label
external predicate label(int id, int key, int str, int num, int num_unit);

// Mapping
external predicate mapping(int id, int memory_start, int memory_limit, int file_offset, int filename, int build_id, boolean has_functions, boolean has_filenames, boolean has_line_numbers, boolean has_inline_frames);

// Location
external predicate location(int id, int mapping_id, int address, boolean is_folded);
external predicate location_to_line(int location_id, int index, int line_id);

// Line
external predicate line(int id, int function_id, int line_number, int column);

// Function
external predicate function(int id, int name, int system_name, int filename, int start_line);

// StringTable
external predicate string_table(int id, string str);
// findStr returns the string associated with the id in string_table
string findStr(int id) {
    string_table(id, result)
}


