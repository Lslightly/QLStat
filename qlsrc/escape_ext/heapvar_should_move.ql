// detect false-sharing problem in https://www.zhihu.com/question/1983394937543366516. These heap allocated object should be moved to goroutine
import go
import heapvar_use


// reference to var is in a closure
predicate refInClosure(RefHeapVar ref) {
    ref.getEnclosingFunction() != ref.getAVar().getDeclaringFunction()
}

from MovedToHeapVar var
where allGoStmtRefsVarInSameScope(var)
    and forall(RefHeapVar ref
        | ref.getAVar() = var and not inGoStmt(ref)
        | not exists(GoStmt gostmt | gostmt = goStmtRefsVar(var) | laterThan(ref, gostmt))
          and not refInClosure(ref) // reference should appear in another closure since capturing the dataflow of closure is hard
    ) // ref which is not in go closure should not be later than gostmt which refs var
    and count(goStmtRefsVar(var)) = 1 // only 1 GoStmt references var. Otherwise, >=2 means sharing.
select var.getLocation() as varDefLoc
