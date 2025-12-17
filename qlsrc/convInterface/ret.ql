/**
 * return to interface type
 */
import go
import lib.typeHelper
import lib.InterfaceLib

from ReturnStmt ret, int i
where
    ret.getEnclosingFunction().getType().getResultType(i) instanceof InterfaceType
select "ret", i as idx, ret.getExpr(i).getType() as srcType, typeSize(ret.getExpr(i).getType()) as srcTypeSize, convSrcTypeSummary(ret.getExpr(i).getType()) as srcKind, ret.getLocation() as loc
