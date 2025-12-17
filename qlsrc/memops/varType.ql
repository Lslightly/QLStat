/**
 * 查看变量类型以及定义位置（排除在类型定义中使用的变量名）
 */
import go
import lib.helper

from Variable var
where not isVarInTypeDef(var)
select var.getType(), var.getDeclaration().getLocation()
