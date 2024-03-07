/**
 * consecutive deref times for patterns like `**p`
 */
import go
import helper

predicate countDerefTimes(Expr expr, int times) {
    if not isExplicitDeref(expr) then
        times = 0
    else
        exists(int subTimes | countDerefTimes(expr.(StarExpr).getBase(), subTimes) | times = 1+subTimes)
}

from StarExpr expr, int times
where isExplicitDeref(expr) and not isExplicitDeref(expr.getParent()) and countDerefTimes(expr, times)
select expr, times, expr.getLocation()
