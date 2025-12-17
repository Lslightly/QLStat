import go
import lib.AliasAnalysis



/**
 * `n` is `x.g`, `m` is `b.f.g`
 * or
 * `n` is `x.g`, `m` is `x.g`
 */
from SelectorExpr n, SelectorExpr m
where isAliasedSelector(n, m)
select n, n.getLocation().getStartLine() as nStart, m, m.getLocation().getStartLine() as mStart

// from SsaVariable ssa
// select ssa, ssa.getAUse()

// from IR::Instruction inst
// select inst