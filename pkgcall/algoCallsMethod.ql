/**
 * 
 * recognize functions which call method or function in specific package many times and consider these functions as algorithms
 * 
 * return enclosing function's name, times the function call specified methods, the location of enclosing function
 * 
 */
import go
import pkgUser

from FuncDef f
where count(MethodCallOfPkg mcall | mcall.getEnclosingFunction() = f | mcall) > 0
select f.getName() as funcName, count(MethodCallOfPkg mcall | mcall.getEnclosingFunction() = f | mcall) as funcUseCnt, f.getLocation() as funcLoc