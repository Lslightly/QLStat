/**
 * 搜索域访问表达式，排除掉了函数
 */
import go
import lib.helper

from ReferenceExpr ref
where isFieldAccess(ref)
select ref, ref.getLocation()