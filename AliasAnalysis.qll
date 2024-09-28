/**
 * Alias analysis library for go in AST and SSA
 */
import go

/*
    ```go
    x := b.f
    print(x.g)
    print(b.f.g)
    ```

    For this program, x.g and b.f.g are alias
*/


/**
 * This predicate is equal to the constraint that  `base.f` is alias to `alias`
 *                                                 |  ↑   |
 *                                                  `path`
 * `alias` is the ssa variable of `x`(the alias of `b.f`)
 * `base` is the ssa variable of `b`
 * `selector` is the selector expression `b.f`
 * `path` is access path string of `b.f`, evaluating to `b.f`
 */
predicate isAliasedThroughAccessPath(SsaVariable alias, SsaVariable base, string path) {
    ( // x(alias) = b(base)
        globalValueNumber(DataFlow::ssaNode(alias)) = globalValueNumber(DataFlow::ssaNode(base))
        and path = base.getSourceVariable().getName()
    )
    or exists(SelectorExpr selector | base.getAUse().isFirstNodeOf(selector.getBase+()) |
        globalValueNumber(DataFlow::ssaNode(alias)) = globalValueNumber(DataFlow::exprNode(selector))
        and getAccessPath(selector) = path
    )
}

/**
 * input `x.f`, output string `x.f`
 * input `b.f.g`, output string `b.f.g`
 */
string getAccessPath(SelectorExpr e) {
    if e.getBase() instanceof SelectorExpr then
        exists(string root | root = getAccessPath(e.getBase())|
        result = root + "." + e.getSelector().getName()
        )
    else 
        result = e.getBase().(VariableName) + "." + e.getSelector().getName()
}

/**
 * `n` is `x.g`
 * `m` is `b.f.g`
 * `baseOfN` is ssa variable of `x`
 * `baseOfM` is ssa variable of `b` in `b.f.g`
 * `path` is the string `b.f`
 * string `b.f.g` = string `x.g`.regexpReplaceAll("^x\\.", "b.f.")
 */
predicate isAliasedSelector(SelectorExpr n, SelectorExpr m) {
    n != m
    and (
        exists(SsaVariable baseOfN, SsaVariable baseOfM, string path |
            isAliasedThroughAccessPath(baseOfN, baseOfM, path) 
            and baseOfN.getAUse().isFirstNodeOf(n.getBase+()) and
            baseOfM.getAUse().isFirstNodeOf(m.getBase+()) and
            getAccessPath(m) = getAccessPath(n).regexpReplaceAll("^"+baseOfN.getSourceVariable().getName()+"\\.", path + ".")
        ) or exists(SsaVariable baseOfN, SsaVariable baseOfM, string path |
            isAliasedThroughAccessPath(baseOfN, baseOfM, path) 
            and baseOfN.getAUse().isFirstNodeOf(m.getBase+()) and
            baseOfM.getAUse().isFirstNodeOf(n.getBase+()) and
            getAccessPath(n) = getAccessPath(m).regexpReplaceAll("^"+baseOfN.getSourceVariable().getName()+"\\.", path + ".")
        )
    )
}

