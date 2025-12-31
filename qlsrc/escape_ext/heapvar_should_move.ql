// detect false-sharing problem in https://www.zhihu.com/question/1983394937543366516. These heap allocated object should be moved to goroutine
import go
import heapvar_use


/**
 * judge by line, which may introduce inaccuracy
 */
predicate laterThan(RefHeapVar ref, GoStmt gostmt) {
    ref.getLocation().getStartLine() > gostmt.getLocation().getEndLine()
}

predicate refNotInGo(RefHeapVar ref) {
    not exists(GoStmt gostmt | ref.getParent*() = gostmt)
}


from MovedToHeapVar var, RefHeapVarInGo refInGo
where allGoStmtRefsVarInSameScope(var)
    and not
        exists(RefHeapVar ref // ref which is not in go closure but later than gostmt which refs var
            | refNotInGo(ref)
              and exists(GoStmt gostmt | gostmt = goStmtRefsVar(var) | laterThan(ref, gostmt)))
    and refInGo.getAVar() = var
select var.getName() as varName, var.getLocation() as varDefLoc, refInGo.getLocation() as refInGoLoc
