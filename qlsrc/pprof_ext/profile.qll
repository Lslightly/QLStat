import profile_ext

// Profile itself is id
class Profile extends int { 
    private int drop_frames;
    private int keep_frames;
    int time_nanos;
    int duration_nanos;
    private int period_type;
    int period;
    private int default_sample_type;
    int doc_url;
    Profile() {
        profile(
            this,
            drop_frames,
            keep_frames,
            time_nanos,
            duration_nanos,
            period_type,
            period,
            default_sample_type,
            doc_url
        )
    }
    ValueType getSampleType(int index) {
        profile_to_sample_type(index, result)
    }
    Sample getSample(int index) {
        profile_to_sample(index, result)
    }
    Mapping getMapping(int index) {
        profile_to_mapping(index, result)
    }
    Location getLocation(int index) {
        profile_to_location(index, result)
    }
    Function getFunction(int index) {
        profile_to_function(index, result)
    }
    string getDropFrames() {
        result = findStr(drop_frames)
    }
    string getKeepFrames() {
        result = findStr(keep_frames)
    }
    ValueType getPeriodType() {
        result = period_type
    }
    string getComment(int index) {
        exists(int comment_id | profile_to_comment(index, comment_id) | result = findStr(comment_id))
    }
    string getDefaultSampleType() {
        result = findStr(default_sample_type)
    }
}

class ValueType extends int {
    private int type;
    private int unit;
    ValueType() {
        value_type(
            this,
            type,
            unit
        )
    }

    string getType() {
        result = findStr(type)
    }
    string getUnit() {
        result = findStr(unit)
    }
}

class Sample extends int {
    Sample() {
        sample_to_location_id(this, _, _)
    }
    Location getLocation(int index) {
        sample_to_location_id(this, index, result)
    }
    int getValue(int index) {
        sample_to_value(this, index, result)
    }
    Label getLabel(int index) {
        sample_to_label(this, index, result)
    }
}

class Label extends int {
    private int key;
    private int str;
    int num;
    private int num_unit;
    Label() {
        label(this, key, str, num, num_unit)
    }
    string getKey() {
        result = findStr(key)
    }
    string getStr() {
        result = findStr(str)
    }
    string getNumUnit() {
        result = findStr(num_unit)
    }
}

class Mapping extends int {
    int memory_start;
    int memory_limit;
    int file_offset;
    private int filename;
    private int build_id;
    boolean has_functions;
    boolean has_filenames;
    boolean has_line_numbers;
    boolean has_inline_frames;
    Mapping() {
        mapping(this, memory_start, memory_limit, file_offset, filename, build_id, has_functions, has_filenames, has_line_numbers, has_inline_frames)
    }
    string getFilename() {
        result = findStr(filename)
    }
    string getBuildId() {
        result = findStr(build_id)
    }
}

class Location extends int {
    private int mapping_id;
    int address;
    boolean is_folded;
    Location() {
        location(this, mapping_id, address, is_folded)
    }
    Mapping getMapping() {
        result = mapping_id
    }
    Line getLine(int index) {
        location_to_line(this, index, result)
    }
}

class Line extends int {
    private int function_id;
    int line_number;
    int column;
    Line() {
        line(this, function_id, line_number, column)
    }
    Function getFunction() {
        result = function_id
    }
}

class Function extends int {
    private int name;
    private int system_name;
    private int filename;
    int start_line;
    Function() {
        function(this, name, system_name, filename, start_line)
    }
    string getName() {
        result = findStr(name)
    }
    string getSystemName() {
        result = findStr(system_name)
    }
    string getFilename() {
        result = findStr(filename)
    }
}

