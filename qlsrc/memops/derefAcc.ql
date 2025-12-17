/**
 * 解引用操作，包括`*a`和`(*x).f`
 */
import go
import lib.helper

from Expr expr
where  isExplicitDeref(expr)
    or isDerefFieldAcc(expr)
select expr, expr.getLocation()
