/**
 * assign to interface with allocation
 * LHS = RHS
 * BoolType, NumericType, StringType, ArrayType, SliceType, StructType, SignatureType, MapType, ChannelType, not InterfaceType, not PointerType
 */
import go
import lib.typeHelper
import lib.InterfaceLib

from Assignment assign, int idx
where
    assign.getLhs(idx).getType().getUnderlyingType() instanceof InterfaceType
    and convSrcTypeSummary(assign.getRhs(idx).getType()) = "pointer"
select "assign idx", idx, assign.getRhs(idx).getType() as srcType, typeSize(assign.getRhs(idx).getType()) as srcTypeSize, assign.getLocation() as loc
