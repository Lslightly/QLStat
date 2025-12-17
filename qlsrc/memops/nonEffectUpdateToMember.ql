import go

from Method m, Variable recv, Write w, Field f
where
  recv = m.getReceiver() and
  w.writesField(recv.getARead(), f, _) and
  not recv.getType() instanceof PointerType
select w, "This update to " + f + " has no effect, because " + recv + " is not a pointer."