import escape_ext
import lib.helper

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

/**
 * for all gostmt which references var in its closure
 * if scopes of these gostmts are same as var's scope, then
 * var should be considered
 */
predicate allGoStmtRefsVarInSameScope(MovedToHeapVar var) {
    forex(GoStmt gostmt
        |gostmt = goStmtRefsVar(var)
        |getEnclosingBlock(gostmt) = getEnclosingBlock(var.getDeclaration())
    )
}

/**
 * gostmt whose closure references var
 */
GoStmt goStmtRefsVar(MovedToHeapVar var) {
    exists(RefHeapVarInGo ref | ref.getAVar() = var | ref.getAWrapGo() = result)
}

/**
 * judge by line, which may introduce inaccuracy
 */
predicate laterThan(RefHeapVar ref, GoStmt gostmt) {
    ref.getLocation().getStartLine() > gostmt.getLocation().getEndLine()
}
