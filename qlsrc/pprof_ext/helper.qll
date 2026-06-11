/**
 * Computes (numerator / denominator) * 100 as a float with `precision` decimal places.
 * Uses fixed-point rounding on BigInt division.
 */
bindingset[numerator, denominator, precision]
float percent(QlBuiltins::BigInt numerator, QlBuiltins::BigInt denominator, int precision) {
    result = (((10.toBigInt().pow(2+precision+1)) * numerator / denominator+5.toBigInt())/10.toBigInt()).toString().toFloat()/10.pow(precision)
}

/**
 * Divides two BigInts and returns a float with `precision` decimal places (rounded).
 */
bindingset[numerator, denominator, precision]
float divBigInt(QlBuiltins::BigInt numerator, QlBuiltins::BigInt denominator, int precision) {
    result = (((10.toBigInt().pow(precision+1)) * numerator / denominator+5.toBigInt())/10.toBigInt()).toString().toFloat()/10.pow(precision)
}

/**
 * funcFullName returns pkgPath+"."+funcName
 */
bindingset[pkgPath, funcName]
string funcFullName(string pkgPath, string funcName) {
    result = pkgPath+"."+funcName
}

/**
 * return funcName part of pkgPath+"."+funcName
 */
bindingset[fullname]
string funcNamePart(string fullname) {
    result = fullname.suffix(fullname.splitAt(".", 0).length()+1)
}
