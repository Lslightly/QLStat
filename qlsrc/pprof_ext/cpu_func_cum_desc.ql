import cpuprofile

from CPUProfile profile, string funcFName
where profile.containsFunc(funcFName)
select funcNamePart(funcFName) as funcName, timeInUnit(profile.cumTimeOfFunc(funcFName), "ms", 4) as cum
order by cum desc
