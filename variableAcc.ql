/**
 * 排除掉函数名称、*a, &a, var a int, a := 1, a[1]的变量访存
 * 
 * TODO 这里实际上对于(x)这种形式的统计存在问题
 */
import go

from ReferenceExpr ref
where not(ref instanceof FunctionName)
    and not (ref.getParent*() instanceof StarExpr)
    and not (ref.getParent*() instanceof SelectorExpr)
    and not (ref.getParent*() instanceof AddressExpr)
    and not (ref.getParent*() instanceof DeclStmt)
    and not (ref.getParent*() instanceof DefineStmt)
    and not (ref.getParent*() instanceof IndexExpr)
select ref, ref.getLocation()
