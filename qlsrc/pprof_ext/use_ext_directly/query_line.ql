import pprof_ext.profile_ext

from string funcName, int lineNumber, string comment
where queryLine(funcName, lineNumber, comment)
select funcName, lineNumber, comment
