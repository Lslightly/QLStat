// import heapvar_use
import go

from ReferenceExpr ref, GoStmt gostmt
where ref.getParent*() = gostmt
select ref, ref.getLocation() as refLoc
