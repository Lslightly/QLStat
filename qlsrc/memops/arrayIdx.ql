/**
 * `a[i]` where `a` is a array
 */
import go
import lib.helper

from ReferenceExpr ref
where ref instanceof IndexExpr and ref.(IndexExpr).getBase().getType() instanceof ArrayType
select ref, ref.getLocation()
