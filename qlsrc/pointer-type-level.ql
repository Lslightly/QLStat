/**
 * 查询变量类型中指针嵌套深度
 */
import go
import helper
import typeHelper

from Variable var
where not isVarInTypeDef(var)
select var.getDeclaration().getLocation(), pointerDepthFn(var.getType(), true)
