import pprof_ext.profile

select count(Sample sample) as sample_cnt, count(Function func) as function_cnt, count(Label label) as label_cnt, count(Line line) as line_cnt, count(Location loc) as location_cnt, count(Mapping map) as mapping_cnt, max(int id | string_table(id, _) |id)+1 as string_table_cnt, count(ValueType vt) as value_type_cnt
