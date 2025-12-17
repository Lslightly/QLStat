/**
 * consecutive field access times for patterns like `a.f1.f2.f3`
 */
import go
import lib.helper

predicate countFieldAccTimes(Expr expr, int times) {
    if not isFieldAccess(expr) then
        times = 0
    else
        exists(int subTimes | countFieldAccTimes(expr.(SelectorExpr).getBase(), subTimes) | times = subTimes+1)
}

from SelectorExpr expr, int times
where isFieldAccess(expr) and not isFieldAccess(expr.getParent())
    and countFieldAccTimes(expr, times)
select expr, times, expr.getLocation()