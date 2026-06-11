import cpuprofile

from CPUProfile p, string rootFuncFName
where rootFuncFName = p.flameGraphRootFuncName()
select rootFuncFName, timeInUnit(p.cumTimeOfFunc(rootFuncFName), "ms", 4) as rootCumTime
order by rootCumTime desc
