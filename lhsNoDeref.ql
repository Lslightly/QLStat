/**
 * LHS of assignment which does not contain deref
 */
import go
import helper

from Assignment assign, int lhsIdx
where not isOrContainsDeref(assign.getLhs(lhsIdx))
select assign, assign.getLhs(lhsIdx).toString(), assign.getLocation(), lhsIdx
