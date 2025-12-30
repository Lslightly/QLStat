import escape_ext

class RefHeapVar extends ReferenceExpr {
    MovedToHeapVar var;
    RefHeapVar() {
        this instanceof Ident and
        this.(Ident).refersTo(var)
    }

    MovedToHeapVar getAVar() { result = var }
}

class RefHeapVarInGo extends RefHeapVar {
    GoStmt wrapGoStmt;
    RefHeapVarInGo() {
        this.getParent*() = wrapGoStmt
    }

    GoStmt getAWrapGo() { result = wrapGoStmt }
}