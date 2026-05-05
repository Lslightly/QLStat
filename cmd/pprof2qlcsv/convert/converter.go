package convert

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/pprof/profile"
)

// stringTableBuilder 重建 pprof 的字符串表。
// github.com/google/pprof/profile 库在解析 protobuf 时已将 string table index 展开为实际字符串，
// 而 profile_ext.qll 中的 predicate 需要通过整数索引引用字符串（由 findStr() 在查询时解析），
// 因此需要收集所有出现过的字符串，为每条字符串分配唯一索引，重建 string table。
type stringTableBuilder struct {
	strings []string
	index   map[string]int
}

func newStringTableBuilder() *stringTableBuilder {
	// pprof 规范要求 string_table[0] 必须为空字符串
	return &stringTableBuilder{
		strings: []string{""},
		index:   map[string]int{"": 0},
	}
}

// add 将字符串加入 table，若已存在则返回已有索引，否则分配新索引。
func (b *stringTableBuilder) add(s string) int {
	if idx, ok := b.index[s]; ok {
		return idx
	}
	idx := len(b.strings)
	b.strings = append(b.strings, s)
	b.index[s] = idx
	return idx
}

// valueTypeRegistry 管理 value_type 实体的 id 分配。
// proto 中 ValueType 被多处引用（SampleType 和 PeriodType），
// PeriodType 不一定出现在 SampleType 中，因此需要统一的注册表来分配 id。
type valueTypeRegistry struct {
	nextID int
	byKey  map[valueTypeKey]int // (type,unit) → id
	rows   [][]string
}

type valueTypeKey struct {
	typ  string
	unit string
}

func newValueTypeRegistry() *valueTypeRegistry {
	return &valueTypeRegistry{
		nextID: 0,
		byKey:  make(map[valueTypeKey]int),
	}
}

// register 将 ValueType 注册到表中，若 (type,unit) 已存在则返回已有 id，否则分配新 id。
func (r *valueTypeRegistry) register(stb *stringTableBuilder, typ, unit string) int {
	key := valueTypeKey{typ, unit}
	if id, ok := r.byKey[key]; ok {
		return id
	}
	id := r.nextID
	r.nextID++
	r.byKey[key] = id
	r.rows = append(r.rows, []string{
		strconv.Itoa(id),
		strconv.Itoa(stb.add(typ)),
		strconv.Itoa(stb.add(unit)),
	})
	return id
}

// lineRegistry 管理 line 实体的 id 分配。
// 不同的 Location 可能引用相同的 (function_id, line_number, column) 组合，
// 通过注册表去重，确保 line 实体表没有重复行。
type lineRegistry struct {
	nextID int
	byKey  map[lineKey]int // (function_id, line_number, column) → id
	rows   [][]string
}

type lineKey struct {
	functionID uint64
	lineNumber int64
	column     int
}

func newLineRegistry() *lineRegistry {
	return &lineRegistry{
		nextID: 0,
		byKey:  make(map[lineKey]int),
	}
}

func (r *lineRegistry) register(funcID uint64, lineNumber int64, column int) int {
	key := lineKey{funcID, lineNumber, column}
	if id, ok := r.byKey[key]; ok {
		return id
	}
	id := r.nextID
	r.nextID++
	r.byKey[key] = id
	r.rows = append(r.rows, []string{
		strconv.Itoa(id),
		strconv.FormatUint(funcID, 10),
		strconv.FormatInt(lineNumber, 10),
		strconv.Itoa(column),
	})
	return id
}

// labelRegistry 管理 label 实体的 id 分配。
// 以 (key, str, num, numUnit) 作为注册键去重，相同的 Label 只在实体表中存一行。
type labelRegistry struct {
	nextID int
	byKey  map[labelKey]int
	rows   [][]string
}

type labelKey struct {
	key     string
	str     string
	num     int64
	numUnit string
}

func newLabelRegistry() *labelRegistry {
	return &labelRegistry{
		nextID: 0,
		byKey:  make(map[labelKey]int),
	}
}

func (r *labelRegistry) register(stb *stringTableBuilder, key, str string, num int64, numUnit string) int {
	k := labelKey{key, str, num, numUnit}
	if id, ok := r.byKey[k]; ok {
		return id
	}
	id := r.nextID
	r.nextID++
	r.byKey[k] = id
	r.rows = append(r.rows, []string{
		strconv.Itoa(id),
		strconv.Itoa(stb.add(key)),
		strconv.Itoa(stb.add(str)),
		strconv.FormatInt(num, 10),
		strconv.Itoa(stb.add(numUnit)),
	})
	return id
}

// ProfileData 保存转换后的所有 predicate 数据行，每个字段对应一个 CSV 表。
// 字段命名采用简写形式，完整 predicate 名通过 PredicateRows() 和 PredicateNames() 映射。
type ProfileData struct {
	profile     [][]string // profile 单例表
	valueType   [][]string // value_type 实体表
	sample      [][]string // sample 实体表
	sample2loc  [][]string // sample_to_location_id 索引表
	sample2val  [][]string // sample_to_value 索引表
	sample2lbl  [][]string // sample_to_label 索引表
	label       [][]string // label 实体表
	mapping     [][]string // mapping 实体表
	location    [][]string // location 实体表
	loc2line    [][]string // location_to_line 索引表
	line        [][]string // line 实体表
	function    [][]string // function 实体表
	stringTable [][]string // string_table 实体表
	p2st        [][]string // profile_to_sample_type 索引表
	p2s         [][]string // profile_to_sample 索引表
	p2m         [][]string // profile_to_mapping 索引表
	p2l         [][]string // profile_to_location 索引表
	p2f         [][]string // profile_to_function 索引表
	p2stbl      [][]string // profile_to_string_table 索引表
	p2c         [][]string // profile_to_comment 索引表
}

// Convert 将 pprof.Profile 转换为 ProfileData。
// 转换规则：
//   - Profile 为单例，id 固定为 0
//   - Sample 在 proto 中没有显式 id，使用其在 Profile.sample 数组中的索引作为 id
//   - Label 在 proto 中是嵌套消息，通过 labelRegistry 按 (key, str, num, numUnit) 去重
//   - 所有字符串字段（如函数名、文件名等）通过 stringTableBuilder 转换为 string table 索引
//   - proto 中的 repeated 字段展开为多行 CSV，每行添加 index 列
//   - Line.column 在 pprof Go 库中目前无导出字段，CSV 中固定输出 0
//   - Line 通过 lineRegistry 按 (function_id, line_number, column) 去重
//   - ValueType 通过 valueTypeRegistry 按 (type, unit) 去重
func Convert(p *profile.Profile) *ProfileData {
	stb := newStringTableBuilder()
	vtr := newValueTypeRegistry()
	lr := newLineRegistry()
	lbr := newLabelRegistry()

	d := &ProfileData{}

	// ValueType：先为 SampleType 注册 id，再注册 PeriodType
	sampleTypeIDs := make([]int, len(p.SampleType))
	for i, vt := range p.SampleType {
		sampleTypeIDs[i] = vtr.register(stb, vt.Type, vt.Unit)
		d.p2st = append(d.p2st, []string{strconv.Itoa(i), strconv.Itoa(sampleTypeIDs[i])})
	}
	var periodTypeID int
	if p.PeriodType != nil {
		periodTypeID = vtr.register(stb, p.PeriodType.Type, p.PeriodType.Unit)
	}
	d.valueType = vtr.rows

	// Profile 单例：id=0，字符串字段存入 string table 取索引
	d.profile = [][]string{{
		"0",
		strconv.Itoa(stb.add(p.DropFrames)),
		strconv.Itoa(stb.add(p.KeepFrames)),
		strconv.FormatInt(p.TimeNanos, 10),
		strconv.FormatInt(p.DurationNanos, 10),
		strconv.Itoa(periodTypeID),
		strconv.FormatInt(p.Period, 10),
		strconv.Itoa(stb.add(p.DefaultSampleType)),
		strconv.Itoa(stb.add(p.DocURL)),
	}}

	// Mapping：id 为 proto 中显式指定的 Mapping.ID
	d.mapping = make([][]string, len(p.Mapping))
	for i, m := range p.Mapping {
		d.mapping[i] = []string{
			strconv.FormatUint(m.ID, 10),
			strconv.FormatUint(uint64(m.Start), 10),
			strconv.FormatUint(uint64(m.Limit), 10),
			strconv.FormatUint(m.Offset, 10),
			strconv.Itoa(stb.add(m.File)),
			strconv.Itoa(stb.add(m.BuildID)),
			strconv.FormatBool(m.HasFunctions),
			strconv.FormatBool(m.HasFilenames),
			strconv.FormatBool(m.HasLineNumbers),
			strconv.FormatBool(m.HasInlineFrames),
		}
		// profile_to_mapping
		d.p2m = append(d.p2m, []string{strconv.Itoa(i), strconv.FormatUint(m.ID, 10)})
	}

	// Location：id 为 proto 中显式指定的 Location.ID
	// 用 locIDSet 去重，因为 p.Location 列表中可能有重复 ID
	locIDSet := make(map[uint64]bool)
	for _, loc := range p.Location {
		if locIDSet[loc.ID] {
			continue
		}
		locIDSet[loc.ID] = true
		mappingID := uint64(0)
		if loc.Mapping != nil {
			mappingID = loc.Mapping.ID
		}
		d.location = append(d.location, []string{
			strconv.FormatUint(loc.ID, 10),
			strconv.FormatUint(mappingID, 10),
			strconv.FormatUint(loc.Address, 10),
			strconv.FormatBool(loc.IsFolded),
		})
		// Location 内嵌的 Line 数组展开，通过 lineRegistry 注册（相同 Line 去重）
		for j, ln := range loc.Line {
			funcID := uint64(0)
			if ln.Function != nil {
				funcID = ln.Function.ID
			}
			currentLineID := lr.register(funcID, ln.Line, 0)
			// location_to_line
			d.loc2line = append(d.loc2line, []string{
				strconv.FormatUint(loc.ID, 10),
				strconv.Itoa(j),
				strconv.Itoa(currentLineID),
			})
		}
	}
	d.line = lr.rows
	for i, loc := range p.Location {
		// profile_to_location
		d.p2l = append(d.p2l, []string{strconv.Itoa(i), strconv.FormatUint(loc.ID, 10)})
	}

	// Function：id 为 proto 中显式指定的 Function.ID
	// 用 funcIDSet 去重，因为 p.Function 列表中可能有重复 ID
	funcIDSet := make(map[uint64]bool)
	for _, f := range p.Function {
		if funcIDSet[f.ID] {
			continue
		}
		funcIDSet[f.ID] = true
		d.function = append(d.function, []string{
			strconv.FormatUint(f.ID, 10),
			strconv.Itoa(stb.add(f.Name)),
			strconv.Itoa(stb.add(f.SystemName)),
			strconv.Itoa(stb.add(f.Filename)),
			strconv.FormatInt(f.StartLine, 10),
		})
	}
	for i, f := range p.Function {
		// profile_to_function
		d.p2f = append(d.p2f, []string{strconv.Itoa(i), strconv.FormatUint(f.ID, 10)})
	}

	// Sample：proto 中无显式 id，使用数组索引作为 sample_id
	for i, s := range p.Sample {
		sampleID := i
		d.sample = append(d.sample, []string{strconv.Itoa(sampleID)})

		// sample_to_location_id：展开 Sample.Location 数组
		for j, loc := range s.Location {
			d.sample2loc = append(d.sample2loc, []string{
				strconv.Itoa(sampleID),
				strconv.Itoa(j),
				strconv.FormatUint(loc.ID, 10),
			})
		}

		// sample_to_value：展开 Sample.Value 数组
		for j, v := range s.Value {
			d.sample2val = append(d.sample2val, []string{
				strconv.Itoa(sampleID),
				strconv.Itoa(j),
				strconv.FormatInt(v, 10),
			})
		}

		// Sample.Label 和 Sample.NumLabel：通过 labelRegistry 注册（相同 Label 去重）
		lblIdx := 0
		// 字符串标签：s.Label map[string][]string
		for key, vals := range s.Label {
			for _, val := range vals {
				currentLabelID := lbr.register(stb, key, val, 0, "")
				d.sample2lbl = append(d.sample2lbl, []string{
					strconv.Itoa(sampleID),
					strconv.Itoa(lblIdx),
					strconv.Itoa(currentLabelID),
				})
				lblIdx++
			}
		}
		// 数值标签：s.NumLabel map[string][]int64，s.NumUnit map[string][]string
		for key, vals := range s.NumLabel {
			units := s.NumUnit[key]
			for j, num := range vals {
				unit := ""
				if j < len(units) {
					unit = units[j]
				}
				currentLabelID := lbr.register(stb, key, "", num, unit)
				d.sample2lbl = append(d.sample2lbl, []string{
					strconv.Itoa(sampleID),
					strconv.Itoa(lblIdx),
					strconv.Itoa(currentLabelID),
				})
				lblIdx++
			}
		}

		// profile_to_sample
		d.p2s = append(d.p2s, []string{strconv.Itoa(i), strconv.Itoa(sampleID)})
	}
	d.label = lbr.rows

	// string_table 实体和 profile_to_string_table 索引
	d.stringTable = make([][]string, len(stb.strings))
	for i, s := range stb.strings {
		d.stringTable[i] = []string{strconv.Itoa(i), s}                         // row: id, string
		d.p2stbl = append(d.p2stbl, []string{strconv.Itoa(i), strconv.Itoa(i)}) // row: index, string_table_id
	}

	// profile_to_comment：展开 Profile.Comments 数组
	for i, c := range p.Comments {
		d.p2c = append(d.p2c, []string{strconv.Itoa(i), strconv.Itoa(stb.add(c))})
	}

	return d
}

// PredicateRows 根据 predicate 名返回对应的 CSV 数据行。
func (d *ProfileData) PredicateRows(name string) [][]string {
	switch name {
	case "profile":
		return d.profile
	case "value_type":
		return d.valueType
	case "sample":
		return d.sample
	case "sample_to_location_id":
		return d.sample2loc
	case "sample_to_value":
		return d.sample2val
	case "sample_to_label":
		return d.sample2lbl
	case "label":
		return d.label
	case "mapping":
		return d.mapping
	case "location":
		return d.location
	case "location_to_line":
		return d.loc2line
	case "line":
		return d.line
	case "function":
		return d.function
	case "string_table":
		return d.stringTable
	case "profile_to_sample_type":
		return d.p2st
	case "profile_to_sample":
		return d.p2s
	case "profile_to_mapping":
		return d.p2m
	case "profile_to_location":
		return d.p2l
	case "profile_to_function":
		return d.p2f
	case "profile_to_string_table":
		return d.p2stbl
	case "profile_to_comment":
		return d.p2c
	default:
		return nil
	}
}

// PredicateNames 返回所有 predicate 的输出顺序。
func (d *ProfileData) PredicateNames() []string {
	predicateOrder := []string{
		"profile",
		"value_type",
		"string_table",
		"mapping",
		"location",
		"location_to_line",
		"line",
		"function",
		"label",
		"sample",
		"sample_to_location_id",
		"sample_to_value",
		"sample_to_label",
		"profile_to_sample_type",
		"profile_to_sample",
		"profile_to_mapping",
		"profile_to_location",
		"profile_to_function",
		"profile_to_string_table",
		"profile_to_comment",
	}
	return predicateOrder
}

// DumpCSV 将所有 predicate 数据写入 CSV 文件到 dir 目录下。
// 每个 predicate 生成一个 <name>.csv 文件，首行为列名（从 predicateSchemas 获取）。
func (d *ProfileData) DumpCSV(dir string) error {
	for _, name := range d.PredicateNames() {
		rows := d.PredicateRows(name)
		header := predicateSchemas[name]
		if header == nil {
			return fmt.Errorf("unknown predicate: %s", name)
		}
		if err := writeCSV(dir, name, rows); err != nil {
			return fmt.Errorf("writing %s: %w", name, err)
		}
	}
	return nil
}

// writeCSV 将 header 和 rows 写入 dir/name.csv 文件。
func writeCSV(dir, name string, rows [][]string) error {
	path := filepath.Join(dir, name+".csv")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return w.Error()
}
