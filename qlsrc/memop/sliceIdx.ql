/**
 * `s[i]` where `s` is slice
 */
import go
import helper

from ReferenceExpr ref
where ref instanceof IndexExpr and ref.(IndexExpr).getBase().getType() instanceof SliceType
select ref, ref.getLocation()
