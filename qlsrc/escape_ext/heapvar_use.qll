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

BlockStmt getEnclosingBlock(Stmt stmt) {
    result = stmt.getParent()
}

/**
 * for all gostmt which references var in its closure
 * if scopes of these gostmts are same as var's scope, then
 * var should be considered
 */
predicate allGoStmtRefsVarInSameScope(MovedToHeapVar var) {
    forex(GoStmt gostmt
        |gostmt = goStmtRefsVar(var)
        |getEnclosingBlock(gostmt).getScope() = var.getScope()
    )
}

/**
 * gostmt whose closure references var
 */
GoStmt goStmtRefsVar(MovedToHeapVar var) {
    exists(RefHeapVarInGo ref | ref.getAVar() = var | ref.getAWrapGo() = result)
}
