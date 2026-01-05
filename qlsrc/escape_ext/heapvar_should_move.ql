// detect false-sharing problem in https://www.zhihu.com/question/1983394937543366516. These heap allocated object should be moved to goroutine
import go
import heapvar_use


// reference to var is in a closure
predicate refInClosure(RefHeapVar ref) {
    ref.getEnclosingFunction() != ref.getAVar().getDeclaringFunction()
}

/**
 * TODO
 *      defer
 *      referenced in another function literal which is not go closure
 */

from MovedToHeapVar var
where allGoStmtRefsVarInSameScope(var)
    and forall(RefHeapVar ref
        | ref.getAVar() = var and not inGoStmt(ref)
        | not exists(GoStmt gostmt | gostmt = goStmtRefsVar(var) | laterThan(ref, gostmt))
          and not refInClosure(ref) // reference should appear in another closure since capturing the dataflow of closure is hard
    ) // ref which is not in go closure should not be later than gostmt which refs var
    // and refInGo.getAVar() = var
select var.getLocation() as varDefLoc/*, refInGo.getLocation() as refInGoLoc*/
