import go

private class ParamsGet extends TaintTracking::FunctionModel, Method {
    ParamsGet() { this.hasQualifiedName("github.com/gin-gonic/gin", "Params", "Get") }

    override predicate hasTaintFlow(FunctionInput inp, FunctionOutput outp) {
        inp.isReceiver() and outp.isResult(0)
    }
}
