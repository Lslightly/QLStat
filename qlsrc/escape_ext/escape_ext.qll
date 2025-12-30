import go
import ext_preds

class MovedToHeapVar extends Variable {
    MovedToHeapVar() {
        movedToHeap(this.getLocation().getFile().getRelativePath(), this.getLocation().getStartLine(), this.getLocation().getStartColumn())
    }
}

