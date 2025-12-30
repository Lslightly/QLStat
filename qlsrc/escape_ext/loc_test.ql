/*
used with test/escape repo
*/
import go

from Variable var
where var.getLocation().getStartLine() = 64
select var.getName() as varName, var.getLocation() as varLoc, var.getLocation().getStartLine() as startLine, var.getLocation().getStartColumn() as startCol, var.getLocation().getFile().getRelativePath() as relativePath
