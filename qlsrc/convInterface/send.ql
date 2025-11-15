/**
 * send consists of implicit conversion
 * `ch <- s` where the type of `ch` is `chan interface{}`
 */
import go
import typeHelper
import InterfaceLib

from SendStmt send
where
    send.getChannel().getType().(ChanType).getElementType() instanceof InterfaceType
select "send", send.getValue().getType() as srcType, typeSize(send.getValue().getType()) as srcTypeSize, convSrcTypeSummary(send.getValue().getType()) as srcKind, send.getLocation() as loc
