import heapvar_use

// from Ident id, MovedToHeapVar var
// where var.getName() = "localCount" and id.uses(var)
// select id.getLocation() as loc

from RefHeapVarInGo ref
select ref.getLocation() as refLoc, ref, ref.getAVar().getName()

