import helper
import profile

/**
 * Extension of Profile providing aggregation methods for heap memory data.
 *
 * Go's Heap Profile contains 4 sample values per sample (in order):
 *   0: alloc_objects (count)  — total allocated objects (Poisson-scaled)
 *   1: alloc_space   (bytes)  — total allocated bytes (Poisson-scaled)
 *   2: inuse_objects (count)  — live objects remaining (Poisson-scaled)
 *   3: inuse_space   (bytes)  — live bytes remaining (Poisson-scaled)
 *
 * All values are already adjusted by Go's Poisson scaling,
 * representing estimated true allocation counts/sizes.
 *
 * Each sample also carries a numeric label "bytes" indicating
 * the average block size (AllocBytes / AllocObjects) for that sample.
 */
class HeapProfile instanceof Profile {
    HeapProfile() {
        // Period type must be ("space", "bytes")
        super.getPeriodType().getType() = "space" and
        super.getPeriodType().getUnit() = "bytes" and
        // Exactly 4 sample types: alloc_objects/count, alloc_space/bytes,
        // inuse_objects/count, inuse_space/bytes
        super.getSampleType(0).getType() = "alloc_objects" and
        super.getSampleType(0).getUnit() = "count" and
        super.getSampleType(1).getType() = "alloc_space" and
        super.getSampleType(1).getUnit() = "bytes" and
        super.getSampleType(2).getType() = "inuse_objects" and
        super.getSampleType(2).getUnit() = "count" and
        super.getSampleType(3).getType() = "inuse_space" and
        super.getSampleType(3).getUnit() = "bytes"
    }

    string toString() {
        result = "HeapProfile " + this.(int).toString()
    }

    // -----------------------------------------------------------------------
    // Sum helpers — total across all samples for each value index
    // -----------------------------------------------------------------------

    /**
     * Sum of sample values at index 0: total allocated objects (count).
     */
    QlBuiltins::BigInt allocObjectsSum() {
        result = sum(Sample sample
            | super.getSample(_) = sample
            | sample.getValue(0)
        )
    }

    /**
     * Sum of sample values at index 1: total allocated space (bytes).
     */
    QlBuiltins::BigInt allocSpaceSum() {
        result = sum(Sample sample
            | super.getSample(_) = sample
            | sample.getValue(1)
        )
    }

    /**
     * Sum of sample values at index 2: total in-use objects (count).
     */
    QlBuiltins::BigInt inuseObjectsSum() {
        result = sum(Sample sample
            | super.getSample(_) = sample
            | sample.getValue(2)
        )
    }

    /**
     * Sum of sample values at index 3: total in-use space (bytes).
     */
    QlBuiltins::BigInt inuseSpaceSum() {
        result = sum(Sample sample
            | super.getSample(_) = sample
            | sample.getValue(3)
        )
    }

    // -----------------------------------------------------------------------
    // Cumulative (total) values per function — across all samples containing funcFullName
    // -----------------------------------------------------------------------

    /**
     * Cumulative allocated objects (count) of samples whose call stack contains `funcFullName`.
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt allocObjectsOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName)
            | sample.getValue(0)
        )
    }

    /**
     * Cumulative allocated space (bytes) of samples whose call stack contains `funcFullName`.
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt allocSpaceOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName)
            | sample.getValue(1)
        )
    }

    /**
     * Cumulative in-use objects (count) of samples whose call stack contains `funcFullName`.
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt inuseObjectsOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName)
            | sample.getValue(2)
        )
    }

    /**
     * Cumulative in-use space (bytes) of samples whose call stack contains `funcFullName`.
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt inuseSpaceOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName)
            | sample.getValue(3)
        )
    }

    // -----------------------------------------------------------------------
    // Flat / self-value per function — only samples where funcFullName is at location[0]
    // -----------------------------------------------------------------------

    /**
     * Flat (self) allocated objects (count) of `funcFullName`: only samples where `funcFullName`
     * is at the top of the stack (location index 0).
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt allocObjectsFlatOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName) and
            sample.getLocation(0).getLine(_).getFunction().getName() = funcFullName
            | sample.getValue(0)
        )
    }

    /**
     * Flat (self) allocated space (bytes) of `funcFullName`: only samples where `funcFullName`
     * is at the top of the stack (location index 0).
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt allocSpaceFlatOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName) and
            sample.getLocation(0).getLine(_).getFunction().getName() = funcFullName
            | sample.getValue(1)
        )
    }

    /**
     * Flat (self) in-use objects (count) of `funcFullName`: only samples where `funcFullName`
     * is at the top of the stack (location index 0).
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt inuseObjectsFlatOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName) and
            sample.getLocation(0).getLine(_).getFunction().getName() = funcFullName
            | sample.getValue(2)
        )
    }

    /**
     * Flat (self) in-use space (bytes) of `funcFullName`: only samples where `funcFullName`
     * is at the top of the stack (location index 0).
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt inuseSpaceFlatOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName) and
            sample.getLocation(0).getLine(_).getFunction().getName() = funcFullName
            | sample.getValue(3)
        )
    }

    // -----------------------------------------------------------------------
    // Percentage helpers
    // -----------------------------------------------------------------------

    /**
     * Percentage of allocated objects for `funcFullName` relative to total allocated objects,
     * with `precision` decimal places.
     */
    bindingset[funcFullName, precision]
    float allocObjectsPercent(string funcFullName, int precision) {
        result = percent(allocObjectsOfFunc(funcFullName), allocObjectsSum(), precision)
    }

    /**
     * Percentage of allocated space for `funcFullName` relative to total allocated space,
     * with `precision` decimal places.
     */
    bindingset[funcFullName, precision]
    float allocSpacePercent(string funcFullName, int precision) {
        result = percent(allocSpaceOfFunc(funcFullName), allocSpaceSum(), precision)
    }

    /**
     * Percentage of in-use objects for `funcFullName` relative to total in-use objects,
     * with `precision` decimal places.
     */
    bindingset[funcFullName, precision]
    float inuseObjectsPercent(string funcFullName, int precision) {
        result = percent(inuseObjectsOfFunc(funcFullName), inuseObjectsSum(), precision)
    }

    /**
     * Percentage of in-use space for `funcFullName` relative to total in-use space,
     * with `precision` decimal places.
     */
    bindingset[funcFullName, precision]
    float inuseSpacePercent(string funcFullName, int precision) {
        result = percent(inuseSpaceOfFunc(funcFullName), inuseSpaceSum(), precision)
    }

    // -----------------------------------------------------------------------
    // Value at specific line
    // -----------------------------------------------------------------------

    /**
     * Cumulative allocated space (bytes) of `funcFullName` at a specific `lineNumber`.
     */
    bindingset[funcFullName, lineNumber]
    QlBuiltins::BigInt allocSpaceOfLine(string funcFullName, int lineNumber) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsLine(funcFullName, lineNumber)
            | sample.getValue(1)
        )
    }

    /**
     * Cumulative in-use space (bytes) of `funcFullName` at a specific `lineNumber`.
     */
    bindingset[funcFullName, lineNumber]
    QlBuiltins::BigInt inuseSpaceOfLine(string funcFullName, int lineNumber) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsLine(funcFullName, lineNumber)
            | sample.getValue(3)
        )
    }

    // -----------------------------------------------------------------------
    // Focus context — cumulative value under a calling function
    // -----------------------------------------------------------------------

    /**
     * Allocated space (bytes) of `funcFullName` in the call context of `focusfuncFullName`.
     * Only samples that contain both functions are counted.
     */
    bindingset[funcFullName, focusfuncFullName]
    QlBuiltins::BigInt allocSpaceUnderFunc(string funcFullName, string focusfuncFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName) and
            sample.containsFunc(focusfuncFullName)
            | sample.getValue(1)
        )
    }

    /**
     * In-use space (bytes) of `funcFullName` in the call context of `focusfuncFullName`.
     * Only samples that contain both functions are counted.
     */
    bindingset[funcFullName, focusfuncFullName]
    QlBuiltins::BigInt inuseSpaceUnderFunc(string funcFullName, string focusfuncFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName) and
            sample.containsFunc(focusfuncFullName)
            | sample.getValue(3)
        )
    }

    // -----------------------------------------------------------------------
    // BlockSize label access
    // -----------------------------------------------------------------------

    /**
     * Average block size (bytes) of samples whose call stack contains `funcFullName`.
     * Block size is derived from the "bytes" label on each heap sample,
     * which represents AllocBytes / AllocObjects per sample. Returns the sum
     * of block sizes across all matching samples.
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt blockSizeSumOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName)
            | getBlockSize(sample)
        )
    }

    /**
     * Returns the "bytes" label value (average block size) of a given heap sample.
     */
    QlBuiltins::BigInt getBlockSize(Sample sample) {
        exists(Label label | sample.getLabel(_) = label and label.getKey() = "bytes" and label.getNum() != 0 |
            result = label.getNum().toBigInt()
        )
    }
}

/**
 * Converts a heap allocation byte value `b` to a float in the given unit,
 * with `precision` decimal places.
 * Supported units: "KB", "MB", "GB", "B".
 */
bindingset[b, unit, precision]
float bytesInUnit(QlBuiltins::BigInt b, string unit, int precision) {
    if unit = "KB" then
        result = divBigInt(b, 1024.toBigInt(), precision)
    else if unit = "MB" then
        result = divBigInt(b, 1024.toBigInt().pow(2), precision)
    else if unit = "GB" then
        result = divBigInt(b, 1024.toBigInt().pow(3), precision)
    else
        result = b.toString().toFloat()
}
