/*
    This ql is not used yet.
    assign to interface, count the source type
*/
import go

from Assignment assign, int idx
where
    assign.getLhs(idx).getType().getUnderlyingType() instanceof InterfaceType
select idx, assign.getRhs(idx).getType().getUnderlyingType().getName() as srcTypeName
