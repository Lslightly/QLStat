/**
 * method call for certain packages
 */
import go
import helper

from SelectorExpr methodCall
where isMethodCall(methodCall) and targetOfMCall(methodCall).getPackage().getPath() in ["math/big", "crypto/sha256", "crypto/rsa", "image/jpeg", "encoding/csv"]
select targetOfMCall(methodCall).getPackage().getPath() as pkgOfCall, methodCall.getLocation() as loc
