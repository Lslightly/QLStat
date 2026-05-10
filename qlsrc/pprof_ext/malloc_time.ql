import profile
import cpuprofile

/**
 * Holds true if the sample's call stack contains runtime.mallocgc (memory allocation).
 */
predicate containsMallocgc(Sample sample) {
    sample.getLocation(_).getLine(_).getFunction().getName() = mallocgc()
}

/**
 * Fully qualified name of the BenchmarkMalloc16 benchmark function.
 */
string benchmarkMalloc16Name() {
    result = "github.com/Lslightly/qlstat/repos/test/malloc_test.BenchmarkMalloc16"
}

/**
 * Name of the Go runtime memory allocation core function.
 */
string mallocgc() {
    result = "runtime.mallocgc"
}

/**
 * Queries memory allocation timing metrics:
 *   - mallocgc_time_ms:      total time of samples containing mallocgc (ms)
 *   - mallocgc_percent:      percentage of total sample time spent in mallocgc
 *   - cumTime:               cumulative time of mallocgc (seconds)
 *   - flatTime:              self time of mallocgc (seconds)
 *   - benchmarkMalloc16Time: cumulative time of BenchmarkMalloc16 (seconds)
 *   - mallocgcin16:          mallocgc time within BenchmarkMalloc16 context (seconds)
 */
from CPUProfile profile
select (sum(QlBuiltins::BigInt time, Sample sample | time = sample.getValue(1) and containsMallocgc(sample) | time)/1000000.toBigInt()).toInt() as mallocgc_time_ms, 
    profile.funcPercent(mallocgc(), 2) as mallocgc_percent,
    timeInUnit(profile.cumTimeOfFunc(mallocgc()), "s", 2) as cumTime,
    timeInUnit(profile.flatTimeOfFunc(mallocgc()), "s", 2) as flatTime,
    timeInUnit(profile.cumTimeOfFunc(benchmarkMalloc16Name()), "s", 2) as benchmarkMalloc16Time,
    timeInUnit(profile.cumTimeUnderFunc(mallocgc(), benchmarkMalloc16Name()), "s", 2) as mallocgcin16
