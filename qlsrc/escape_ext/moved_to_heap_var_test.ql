import escape_ext

from MovedToHeapVar var
select var.getName() as varName, var.getLocation() as varLoc, var.getLocation().getStartLine() as startLine, var.getLocation().getStartColumn() as startCol, var.getLocation().getFile().getRelativePath() as relativePath
