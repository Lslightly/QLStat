/**
 * 赋值语句的lhs中包含解引用操作
 */
import go
import lib.helper

from Assignment assign, int lhsIdx
where isOrContainsDeref(assign.getLhs(lhsIdx))
select assign, assign.getLhs(lhsIdx).toString(), assign.getLocation(), lhsIdx