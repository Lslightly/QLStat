/**
 * explicit type conversion expr
 */
import go
import InterfaceLib
import typeHelper

from ConversionExpr conv
where
    conv.getTypeExpr().getType().getUnderlyingType() instanceof InterfaceType
select "conv", conv.getTypeExpr().getType() as tgtType, conv.getOperand().getType() as srcType, typeSize(conv.getOperand().getType()) as srcTypeSize, convSrcTypeSummary(conv.getOperand().getType()) as srcKind, conv.getLocation() as loc
