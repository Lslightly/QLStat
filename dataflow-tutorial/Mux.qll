import go

class RequestVars extends DataFlow::UntrustedFlowSource::Range, DataFlow::CallNode {
    RequestVars() {
        this.getTarget().hasQualifiedName("github.com/gorilla/mux", "Vars")
    }
}

