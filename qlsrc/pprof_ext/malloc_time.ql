import profile

predicate containsMallocgc(Sample sample) {
    sample.getLocation(_).getLine(_).getFunction().getName() = "runtime.mallocgc"
}

select (sum(QlBuiltins::BigInt time, Sample sample | time = sample.getValue(1) and containsMallocgc(sample) | time)/1000000.toBigInt()).toInt() as mallocgc_time_ms
