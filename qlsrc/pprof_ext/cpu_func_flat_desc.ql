import cpuprofile

from CPUProfile profile, string funcFName
where profile.containsFunc(funcFName)
select funcNamePart(funcFName) as funcName, timeInUnit(profile.flatTimeOfFunc(funcFName), "ms", 4) as flat
order by flat desc
