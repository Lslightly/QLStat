/**
 * passing argument to call statement
 */
import go
import typeHelper
import InterfaceLib

from CallExpr call, int idx
where call.getCalleeType().getParameterType(idx).getUnderlyingType() instanceof InterfaceType
    or 
    (call.hasImplicitVarargs() and call.getCalleeType().getParameterType(call.getCalleeType().getNumParameter()-1) instanceof SliceType and call.getCalleeType().getParameterType(call.getCalleeType().getNumParameter()-1).(SliceType).getElementType() instanceof InterfaceType and idx > call.getCalleeType().getNumParameter()-2)
select "call", idx, call.getArgument(idx).getType() as srcType, typeSize(call.getArgument(idx).getType()) as srcTypeSize, convSrcTypeSummary(call.getArgument(idx).getType()) as srcKind, call.getLocation() as loc
