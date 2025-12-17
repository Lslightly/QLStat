/**
 * 查询变量类型中指针嵌套深度
 */
import go
import lib.helper
import lib.typeHelper

from Variable var
where not isVarInTypeDef(var)
select var.getDeclaration().getLocation(), pointerDepthFn(var.getType(), true)
