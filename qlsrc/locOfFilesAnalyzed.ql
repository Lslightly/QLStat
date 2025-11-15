/**
 * 
 * lines of code of files successfully analyzed
 * 
 */
import go

select sum(File f| not exists(Error e | e.getFile() = f) and exists(f.getRelativePath()) | f.getNumberOfLinesOfCode())