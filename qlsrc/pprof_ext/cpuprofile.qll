import helper
import profile

/**
 * Converts nanosecond time `t` to a float in the given unit, with `precision` decimal places.
 * Supported units: "ms" (milliseconds), "s" (seconds), "us" (microseconds).
 * For any other unit, returns the raw value as a float.
 */
bindingset[t, unit, precision]
float timeInUnit(QlBuiltins::BigInt t, string unit, int precision) {
    if unit = "ms" then
        result = divBigInt(t, 10.toBigInt().pow(6), precision)
    else if unit = "s" then
        result = divBigInt(t, 10.toBigInt().pow(9), precision)
    else if unit = "us" then
        result = divBigInt(t, 10.toBigInt().pow(3), precision)
    else
        result = t.toString().toFloat()
}

/**
 * Extension of Profile providing aggregation methods for CPU sampling data.
 */
class CPUProfile extends Profile {
    CPUProfile() {
        super.getPeriodType().getType() = "cpu" and 
        super.getPeriodType().getUnit() = "nanoseconds" and
        super.getSampleType(0).getType() = "samples" and
        super.getSampleType(0).getUnit() = "count" and
        super.getSampleType(1).getType() = "cpu" and 
        super.getSampleType(1).getUnit() = "nanoseconds"
    }

    /**
     * Sum of all sample values at index 1 (total time in nanoseconds).
     */
    QlBuiltins::BigInt sampleTimeSum() {
        result = sum(Sample sample
            | super.getSample(_) = sample
            | sample.getValue(1)
        )
    }

    /**
     * Cumulative time (nanoseconds) of samples whose call stack contains `funcFullName`.
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt cumTimeOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName)
            | sample.getValue(1) // time is at index 1
        )
    }


    /**
     * Flat/self time (nanoseconds) of `funcFullName`: only samples where `funcFullName` is
     * at the top of the stack (location index 0).
     */
    bindingset[funcFullName]
    QlBuiltins::BigInt flatTimeOfFunc(string funcFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName) and
            sample.getLocation(0).getLine(_).getFunction().getName() = funcFullName
            | sample.getValue(1) // time is at index 1
        )
    }

    /**
     * Cumulative time (nanoseconds) of `funcFullName` at a specific `lineNumber`.
     */
    bindingset[funcFullName, lineNumber]
    QlBuiltins::BigInt cumTimeOfLine(string funcFullName, int lineNumber) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsLine(funcFullName, lineNumber)
            | sample.getValue(1) // time is at index 1
        )
    }

    /**
     * Percentage of cumulative time for `funcFullName` relative to total sample time,
     * with `precision` decimal places.
     */
    bindingset[funcFullName, precision]
    float funcPercent(string funcFullName, int precision) {
        result = percent(cumTimeOfFunc(funcFullName), sampleTimeSum(), precision)
    }

    /**
     * Cumulative time (nanoseconds) of `funcFullName` in the call context of `focusfuncFullName`.
     * Only samples that contain both functions are counted.
     */
    bindingset[funcFullName, focusfuncFullName]
    QlBuiltins::BigInt cumTimeUnderFunc(string funcFullName, string focusfuncFullName) {
        result = sum(Sample sample
            | super.getSample(_) = sample and
            sample.containsFunc(funcFullName) and
            sample.containsFunc(focusfuncFullName)
            | sample.getValue(1)
        )
    }

    string flameGraphRootFuncName() {
        exists(Sample sample
            | sample.getLocation(sample.locationNum()-1).getLastLine().getFunction().getName() = result
        )
    }

    predicate containsFunc(string funcFullName) {
        exists(Sample sample |
            sample.containsFunc(funcFullName)
        )
    }
}

