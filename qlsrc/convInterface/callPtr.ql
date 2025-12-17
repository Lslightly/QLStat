/**
 * passing argument to call statement
 */
import go
import lib.typeHelper
import lib.InterfaceLib

from CallExpr call, int idx
where (call.getCalleeType().getParameterType(idx).getUnderlyingType() instanceof InterfaceType
    or 
    (call.hasImplicitVarargs() and call.getCalleeType().getParameterType(call.getCalleeType().getNumParameter()-1) instanceof SliceType and call.getCalleeType().getParameterType(call.getCalleeType().getNumParameter()-1).(SliceType).getElementType() instanceof InterfaceType and idx > call.getCalleeType().getNumParameter()-2))
and convSrcTypeSummary(call.getArgument(idx).getType()) = "pointer"
select "call", idx, call.getArgument(idx).getType() as srcType, typeSize(call.getArgument(idx).getType()) as srcTypeSize, call.getLocation() as loc
