/**
 * goref profile library
 * 
 * Goref is a Go heap object reference analysis tool based on delve.
 * See https://github.com/cloudwego/goref
 * 
 * https://github.com/cloudwego/goref/blob/main/pkg/proc/protobuf.go
 * contains the structure of profile
 */
import profile

class GorefProfile extends Profile {
    GorefProfile() {
        super.getSampleType(0).getType() = "inuse_objects" and
        super.getSampleType(0).getUnit() = "count" and
        super.getSampleType(1).getType() = "inuse_space" and
        super.getSampleType(1).getUnit() = "bytes"
    }
    /** TODO more useful APIs */
}
