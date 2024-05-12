/**
 * package user helper library
 * 
 * package users include call of method and functions defined in package
 */
import go
import helper

class MethodCallOfPkg extends SelectorExpr {
    MethodCallOfPkg() {
        isMethodCall(this) and targetOfMCall(this).getPackage().getPath() in ["math/big", "crypto/sha256", "crypto/rsa", "image/jpeg", "encoding/csv"]
    }

    Function method() {
        result = this.getSelector().(FunctionName).getTarget()
    }

    Package pkg() {
        result = method().getPackage()
    }

    Expr recv() {
        result = this.getBase()
    }

    Type recvType() {
        result = this.getBase().getType()
    }

    Type signature() {
        result = method().getType()
    }

}