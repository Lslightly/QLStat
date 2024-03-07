/**
 * @name Lines of code in repo
 */

import go

select sum(File f| | f.getNumberOfLinesOfCode()) as s
 