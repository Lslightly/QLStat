import escape_ext
import lib.helper

predicate isSmallLoop(LoopStmt loop) {
    loop.getBody().getNumChild() < 5
}

predicate loopContainsNew(LoopStmt loop, NewExprEscapesToHeap new) {
    new.getParent*() = loop
}

predicate loopContainsAppend(LoopStmt loop, AppendExpr append) {
    append.getParent*() = loop
}

predicate loopWithoutGoStmt(LoopStmt loop) {
    not exists(GoStmt gostmt | gostmt.getParent*() = loop)
}

from NewExprEscapesToHeap new, LoopStmt loop
where loopContainsNew(loop, new) and loopWithoutGoStmt(loop)
select new.getLocation() as newLoc, loop.getBody().getNumChildStmt() as numStmtOfLoop
