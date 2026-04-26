package convert

// predicateSchemas 定义 profile_ext.qll 中每个 predicate 对应的 CSV 列名和顺序。
// key 为 predicate 名称（同时作为 CSV 文件名），value 为列名数组，顺序与 qll 中参数顺序一致。
var predicateSchemas = map[string][]string{
	"profile":                 {"id", "drop_frames", "keep_frames", "time_nanos", "duration_nanos", "period_type", "period", "default_sample_type", "doc_url"},
	"value_type":              {"id", "type", "unit"},
	"sample_to_location_id":   {"sample_id", "index", "location_id"},
	"sample_to_value":         {"sample_id", "index", "value"},
	"sample_to_label":         {"sample_id", "index", "label_id"},
	"label":                   {"id", "key", "str", "num", "num_unit"},
	"mapping":                 {"id", "memory_start", "memory_limit", "file_offset", "filename", "build_id", "has_functions", "has_filenames", "has_line_numbers", "has_inline_frames"},
	"location":                {"id", "mapping_id", "address", "is_folded"},
	"location_to_line":        {"location_id", "index", "line_id"},
	"line":                    {"id", "function_id", "line_number", "column"},
	"function":                {"id", "name", "system_name", "filename", "start_line"},
	"string_table":            {"id", "str"},
	"profile_to_sample_type":  {"index", "sample_type_id"},
	"profile_to_sample":       {"index", "sample_id"},
	"profile_to_mapping":      {"index", "mapping_id"},
	"profile_to_location":     {"index", "location_id"},
	"profile_to_function":     {"index", "function_id"},
	"profile_to_string_table": {"index", "string_table_id"},
	"profile_to_comment":      {"index", "comment_id"},
}
