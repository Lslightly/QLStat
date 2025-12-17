/**
 * 
 * pointer assigned to interface
 * 
 */
import go
import lib.helper
import lib.typeHelper

from Assignment assign, int idx
where assign.getLhs(idx).getType().getUnderlyingType() instanceof InterfaceType
    and assign.getRhs(idx).getType().getUnderlyingType() instanceof PointerType
select assign.getLhs(idx).getType() as lhsType, assign.getRhs(idx).getType() as rhsType, typeSize(assign.getRhs(idx).getType()) as rhsTypeSize, assign.getLocation() as loc, idx
