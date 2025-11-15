/**
 * `m[i]` where `m` is a map
 */
import go
import helper

from ReferenceExpr ref
where ref instanceof IndexExpr and ref.(IndexExpr).getBase().getType() instanceof MapType
select ref, ref.getLocation()
