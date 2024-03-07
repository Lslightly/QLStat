/**
 * 统计是VariableName的ReferenceExpr的内存访问
 * 
 */
import go

from ReferenceExpr ref
where ref instanceof VariableName
select count(ref) as cntOfMemAcc
