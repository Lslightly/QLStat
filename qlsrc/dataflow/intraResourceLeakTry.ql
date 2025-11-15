/**
 * intra resource leak(forget to close file)
 */
import go
import semmle.go.dataflow.internal.TaintTrackingUtil

module Config implements DataFlow::ConfigSig {
    predicate isSource(DataFlow::Node src) {
        exists(Function f, DataFlow::CallNode cn | cn = f.getACall() |
            f.hasQualifiedName("os", "Open") and
            src = cn.getResult(0)
        )
    }

    predicate isSink(DataFlow::Node sink) {
        exists(Function f, DataFlow::CallNode cn | cn = f.getACall() |
            f.hasQualifiedName("os.File", "Close") and
            sink = cn.getReceiver()
        )
    }
}

predicate isAlias(DataFlow::Node node1, DataFlow::Node node2) {
    node1.getEnclosingCallable() = node2.getEnclosingCallable() and
    (
        node1 = node2 or
        DataFlow::localFlow(node1, node2)
    )
}

predicate enclosingStmt(Expr e, Stmt stmt) {
    stmt.getAChildExpr+() = e
}

Stmt getEnclosingStmt(Expr e) {
    enclosingStmt(e, result)
}

predicate belongToSameStmt(DataFlow::Node src, DataFlow::Node sink) {
    exists(Expr e1, Expr e2 | e1 = src.asExpr() and e2 = sink.asExpr() and getEnclosingStmt(e1) = getEnclosingStmt(e2))
}

predicate checkDataFlow(DataFlow::Node src, DataFlow::Node sink) {
    Config::isSource(src) and Config::isSink(sink) and src.getEnclosingCallable() = sink.getEnclosingCallable() and (isAlias(src, sink) or belongToSameStmt(src, sink))
}

predicate notClosed(DataFlow::Node src, ControlFlow::Node cfn) {
        (Config::isSource(src) and cfn = src.asInstruction())
    or  (
            notClosed(src, cfn.getAPredecessor())
        and
            not exists(DataFlow::Node sink | sink.asInstruction() = cfn and checkDataFlow(src, sink))
        )
}

module ResourceCloseFlow = TaintTracking::Global<Config>;

from DataFlow::Node src
where notClosed(src, src.getRoot().getExitNode())
select src, src.getStartLine() as startLine, src.getEndLine() as endLine
