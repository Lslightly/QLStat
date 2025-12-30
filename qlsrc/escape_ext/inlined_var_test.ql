import escape_ext

from InlinedMovedToHeapVar call
select call.getCalleeName() as calleeName, call.getLocation() as callLoc, call.getCalleeExpr().getLocation().getEndColumn()+1 as lparen, call.getLocation().getStartColumn() as startCol, call.getLocation().getFile().getRelativePath() as relativePath
