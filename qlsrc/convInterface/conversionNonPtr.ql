/**
 * explicit type conversion expr
 */
import go
import InterfaceLib
import typeHelper

from ConversionExpr conv
where
    conv.getTypeExpr().getType().getUnderlyingType() instanceof InterfaceType
and convSrcTypeSummary(conv.getOperand().getType()) = "alloc"
select "conv", conv.getTypeExpr().getType() as tgtType, conv.getOperand().getType() as srcType, typeSize(conv.getOperand().getType()) as srcTypeSize, conv.getLocation() as loc
