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
        any()
    }

    /**
     * Sum of all sample values at index 1 (total time in nanoseconds).
     */
    QlBuiltins::BigInt sampleTimeSum() {
        result = sum(Sample sample
            | this.getSample(_) = sample
            | sample.getValue(1)
        )
    }

    /**
     * Cumulative time (nanoseconds) of samples whose call stack contains `funcName`.
     */
    bindingset[funcName]
    QlBuiltins::BigInt cumTimeOfFunc(string funcName) {
        result = sum(Sample sample
            | this.getSample(_) = sample and
            sample.containsFunc(funcName)
            | sample.getValue(1) // time is at index 1
        )
    }


    /**
     * Flat/self time (nanoseconds) of `funcName`: only samples where `funcName` is
     * at the top of the stack (location index 0).
     */
    bindingset[funcName]
    QlBuiltins::BigInt flatTimeOfFunc(string funcName) {
        result = sum(Sample sample
            | this.getSample(_) = sample and
            sample.containsFunc(funcName) and
            sample.getLocation(0).getLine(_).getFunction().getName() = funcName
            | sample.getValue(1) // time is at index 1
        )
    }

    /**
     * Cumulative time (nanoseconds) of `funcName` at a specific `lineNumber`.
     */
    bindingset[funcName, lineNumber]
    QlBuiltins::BigInt cumTimeOfLine(string funcName, int lineNumber) {
        result = sum(Sample sample
            | this.getSample(_) = sample and
            sample.containsLine(funcName, lineNumber)
            | sample.getValue(1) // time is at index 1
        )
    }

    /**
     * Percentage of cumulative time for `funcName` relative to total sample time,
     * with `precision` decimal places.
     */
    bindingset[funcName, precision]
    float funcPercent(string funcName, int precision) {
        result = percent(cumTimeOfFunc(funcName), sampleTimeSum(), precision)
    }

    /**
     * Cumulative time (nanoseconds) of `funcName` in the call context of `focusFuncName`.
     * Only samples that contain both functions are counted.
     */
    bindingset[funcName, focusFuncName]
    QlBuiltins::BigInt cumTimeUnderFunc(string funcName, string focusFuncName) {
        result = sum(Sample sample
            | this.getSample(_) = sample and
            sample.containsFunc(funcName) and
            sample.containsFunc(focusFuncName)
            | sample.getValue(1)
        )
    }
}

