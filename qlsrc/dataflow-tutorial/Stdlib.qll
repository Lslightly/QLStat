import go

private class PrintfCall extends LoggerCall::Range, DataFlow::CallNode {
    PrintfCall() {
        this.getTarget().hasQualifiedName("fmt", ["Print", "Printf", "Println"])
    }

    override DataFlow::Node getAMessageComponent() { result = this.getAnArgument() }
}