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

// CallExpr new()'s startCol is "n", while escape analysis's output startCol is "("
predicate locNewEscapesToHeap(Location loc) {
    newEscapesToHeap(loc.getFile().getAbsolutePath(), loc.getStartLine(), loc.getStartColumn()+3)
}

class NewExprEscapesToHeap extends CallExpr {
    NewExprEscapesToHeap() {
        this.getCalleeName() = "new"
        and locNewEscapesToHeap(this.getLocation())
    }
}