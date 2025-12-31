import go
import ext_preds

predicate locMovedToHeap(Location loc) {
    movedToHeap(loc.getFile().getAbsolutePath(), loc.getStartLine(), loc.getStartColumn())
}

class MovedToHeapVar extends LocalVariable {
    MovedToHeapVar() {
        locMovedToHeap(this.getLocation())
    }
}

class InlinedMovedToHeapVar extends CallExpr {
    InlinedMovedToHeapVar() {
        movedToHeap(
            this.getFile().getAbsolutePath(),
            this.getLocation().getStartLine(),
            this.getCalleeExpr().getLocation().getEndColumn()+1 // left parenthesis start col, the location of inlined return var
        )
    }
}
