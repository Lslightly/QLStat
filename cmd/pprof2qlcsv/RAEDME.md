# pprof2qlcsv

工作流：
1. 先用github.com/google/pprof库将解析cpu profile
2. 将解析得到的Go数据按照 [profile_ext.qll](../../qlsrc/pprof_ext/profile_ext.qll) 数据表定义导出csv文件
  1. 在导出数组时，需要添加index编号
3. 在yaml文件中配置`externals`选项，类似于[demo.yaml](../../demo.yaml)

```
cmd/pprof2qlcsv/
├── main.go          # CLI 入口：接收 pprof 文件，输出 CSV 目录
├── converter.go     # 核心转换逻辑：pprof.Profile → CSV rows
└── schema.go        # 从 profile_ext.qll schema 派生的常量定义
                      # （列名、列顺序等，确保和 .qll 一致）
```

## pprof Profile → CSV 映射

参考 [profile.proto](../../doc/adapters/pprof/profile.proto) 中的 protobuf 定义，将 pprof 数据转换为 CodeQL external predicate CSV 文件。

### 映射规则

每个 `profile_ext.qll` 中的 predicate 对应一个 CSV 文件，CSV 文件名与 predicate 名相同。CSV 第一行为列名（与 predicate 参数名一一对应），后续行为数据。

#### 核心实体表

| CSV 文件名 | 对应 predicate | 对应 proto 消息 | 列 |
|---|---|---|---|
| profile.csv | `profile(id, drop_frames, keep_frames, time_nanos, duration_nanos, period_type, period, default_sample_type, doc_url)` | Profile | id,drop_frames,keep_frames,time_nanos,duration_nanos,period_type,period,default_sample_type,doc_url |
| value_type.csv | `value_type(id, type, unit)` | ValueType | id,type,unit |
| string_table.csv | `string_table(id, str)` | Profile.string_table | id,str |
| mapping.csv | `mapping(id, memory_start, memory_limit, file_offset, filename, build_id, has_functions, has_filenames, has_line_numbers, has_inline_frames)` | Mapping | id,memory_start,memory_limit,file_offset,filename,build_id,has_functions,has_filenames,has_line_numbers,has_inline_frames |
| location.csv | `location(id, mapping_id, address, is_folded)` | Location | id,mapping_id,address,is_folded |
| line.csv | `line(id, function_id, line_number, column)` | Line | id,function_id,line_number,column |
| function.csv | `function(id, name, system_name, filename, start_line)` | Function | id,name,system_name,filename,start_line |
| label.csv | `label(id, key, str, num, num_unit)` | Label | id,key,str,num,num_unit |

#### 索引关系表（数组展开，添加 index 列）

| CSV 文件名 | 对应 predicate | 对应 proto 字段 | 列 |
|---|---|---|---|
| profile_to_sample_type.csv | `profile_to_sample_type(index, sample_type_id)` | Profile.sample_type[index] | index,sample_type_id |
| profile_to_sample.csv | `profile_to_sample(index, sample_id)` | Profile.sample[index] | index,sample_id |
| profile_to_mapping.csv | `profile_to_mapping(index, mapping_id)` | Profile.mapping[index] | index,mapping_id |
| profile_to_location.csv | `profile_to_location(index, location_id)` | Profile.location[index] | index,location_id |
| profile_to_function.csv | `profile_to_function(index, function_id)` | Profile.function[index] | index,function_id |
| profile_to_string_table.csv | `profile_to_string_table(index, string_table_id)` | Profile.string_table[index] | index,string_table_id |
| profile_to_comment.csv | `profile_to_comment(index, comment_id)` | Profile.comment[index] | index,comment_id |
| sample_to_location_id.csv | `sample_to_location_id(sample_id, index, location_id)` | Sample.location_id[index] | sample_id,index,location_id |
| sample_to_value.csv | `sample_to_value(sample_id, index, value)` | Sample.value[index] | sample_id,index,value |
| sample_to_label.csv | `sample_to_label(sample_id, index, label_id)` | Sample.label[index] | sample_id,index,label_id |
| location_to_line.csv | `location_to_line(location_id, index, line_id)` | Location.line[index] | location_id,index,line_id |

### 数据类型映射

| proto 类型 | CSV 中表示 | 说明 |
|---|---|---|
| int64/uint64 | 十进制整数字符串 | 如 `"12345"` |
| string | 原始字符串 | string_table 中直接存储字符串值 |
| bool | `"true"` / `"false"` | 小写 |
| int64 (string table index) | 十进制整数字符串 | 存索引值，由 qll 中 `findStr()` 解析 |

### 特殊处理

1. **Profile 是单例**：id 固定为 0
2. **string_table**：proto 中 `repeated string string_table` 是字符串数组，转换时 `id` = 数组索引，`str` = 字符串值。`string_table[0]` 必须为空字符串 `""`
3. **ID 映射**：proto 中 Location/Mapping/Function 的 `id` 字段是显式指定的；而 Sample 在 proto 中没有 id 字段，转换时用其在 `Profile.sample` 数组中的索引作为 id
4. **数组展开**：proto 中的 `repeated` 字段在 CSV 中展开为多行，每行添加 `index` 列记录元素在数组中的位置
5. **ValueType 引用**：`profile.period_type` 是单个 ValueType，在 `profile.csv` 中存该 ValueType 的 id
