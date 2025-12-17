/**
 * return to interface type
 */
import go
import lib.typeHelper
import lib.InterfaceLib

from ReturnStmt ret, int i
where
    ret.getEnclosingFunction().getType().getResultType(i) instanceof InterfaceType
and convSrcTypeSummary(ret.getExpr(i).getType()) = "pointer"
select "ret", i as idx, ret.getExpr(i).getType() as srcType, typeSize(ret.getExpr(i).getType()) as srcTypeSize, ret.getLocation() as loc
