/**
 * go语言查询的帮助函数库
 */
import go

predicate isExplicitDeref(Expr expr) {
    expr instanceof StarExpr and (not expr instanceof TypeExpr)
}

// is the form of `x.f` where `f` is a field instead of a method call
predicate isFieldAccess(Expr expr) {
    expr instanceof SelectorExpr and not (expr.(SelectorExpr).getSelector() instanceof FunctionName)
}

predicate isMethodCall(Expr expr) {
    expr instanceof SelectorExpr and (expr.(SelectorExpr).getSelector() instanceof FunctionName)
}

Function targetOfMCall(SelectorExpr mcall) {
    result = mcall.getSelector().(FunctionName).getTarget()
}

/**
 * x.f is short hand for (*x).f
 */
predicate isImplicitDerefSelector(Expr expr) {
    expr instanceof SelectorExpr and expr.(SelectorExpr).getBase().getType() instanceof PointerType
}

predicate isDerefMethodCall(Expr expr) {
    isImplicitDerefSelector(expr) and expr.(SelectorExpr).getSelector() instanceof FunctionName
}

predicate isDerefFieldAcc(Expr expr) {
    isImplicitDerefSelector(expr) and not isDerefMethodCall(expr)
}

/**
 * Detect whether `expr` is or contains dereference pattern
 */
predicate isOrContainsDeref(Expr expr) {
    isDerefFieldAcc(expr.getAChild*()) or isExplicitDeref(expr.getAChild*())
}

predicate isInTypeDeclOrTypeDefSpec(Expr expr) {
    (expr.getParent*() instanceof TypeDecl)
    or expr.getParent*() instanceof TypeDefSpec
}


/**
 * judge whether the declaration of var is in type definition
 */
predicate isVarInTypeDef(Variable var) {
    isInTypeDeclOrTypeDefSpec(var.getDeclaration())
}

/** Gets the innermost block statement to which this AST node belongs, if any. */
pragma[nomagic]
BlockStmt getEnclosingBlock(AstNode ident) {
    result = parentInSameBlock*(ident.getParent())
}

/** Gets the parent node of this AST node, but without crossing block boundaries. */
private AstNode parentInSameBlock(AstNode node) {
    result = node.getParent()
    and not node instanceof BlockStmt
}