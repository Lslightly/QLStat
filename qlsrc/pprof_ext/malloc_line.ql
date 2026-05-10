import cpuprofile

/**
 * Queries cumulative time at a specific function and line number.
 * Input parameters (funcName, lineNumber, comment) are provided via the external
 * predicate `queryLine`. Outputs the filename:line and cumulative time in milliseconds
 * with 2 decimal places.
 */
from Sample sample, Line line, string funcName, int lineNumber, string comment, CPUProfile profile
where queryLine(funcName, lineNumber, comment) and sample.containsLine(funcName, lineNumber) and // sample contains the query line in its stacktrace
    line.getFunction().getName() = funcName and line.getLineNumber() = lineNumber and
    profile.getSample(_) = sample
select line.getFunction().getFilename()+":"+lineNumber.toString() as loc, timeInUnit(profile.cumTimeOfLine(funcName, lineNumber), "ms", 2) as cumTime
