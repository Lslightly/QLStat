/**
 * method call of packages
 * 
 * return method call location, method name, method package path, method enclosing function location
 */
import pkgUser

from MethodCallOfPkg mcall
select mcall.getLocation() as loc, mcall.method().getName() as methodName, mcall.pkg().getPath() as pkgPath, mcall.getEnclosingFunction().getLocation() as funcLoc