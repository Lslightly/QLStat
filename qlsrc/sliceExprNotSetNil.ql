import go
import semmle.go.Util
import DataFlow
import AliasAnalysis

boolean isTypeSizeLarge(Type t) {
    if not (t instanceof BasicType
        or t instanceof TypeParamType
        or t instanceof TypeSetLiteralType
        or t instanceof PointerType
    ) then
        result = true
    else
        result = false
}

/**
 * s[i:] or s[1:]
 */
predicate isCutFrontSliceExpr(SliceExpr sliceExpr) {
    exists(Expr low |
        low = sliceExpr.getLow()
        and (
            not low.isConst()
            or not low.getExactValue() = "0"
        )
    )
}

/**
 * s = s[:]
 */
predicate selfAssignSliceExpr(Assignment asmt, ReferenceExpr resSlice, SliceExpr sliceExpr) {
    asmt.assigns(resSlice, sliceExpr) and resSlice.toString() = sliceExpr.getBase().toString()
}

/**
 * t = []*objT
 */
predicate isSliceOfPtrType(Type t, Type objT) {
    t.getUnderlyingType() instanceof SliceType
    and t.getUnderlyingType().(SliceType).getElementType().getUnderlyingType() instanceof PointerType
    and t.getUnderlyingType().(SliceType).getElementType().getUnderlyingType().(PointerType).getBaseType() = objT
}

predicate inSameBlock(Stmt s1, Stmt s2) {
    s1.getParent() = s2.getParent()
}

predicate nilAssignToSliceIdx(IndexExpr sliceIdx, Assignment nilAssign) {
    sliceIdx.getBase().getType().getUnderlyingType() instanceof SliceType and
    nilAssign.assigns(sliceIdx, Builtin::nil().getAReference())
}

/**
 * `s1`, `s2` are actually uses of `sliceDef`
 */
predicate useSameSSASlice(Expr s1, Expr s2) {
    exists(
        SsaVariable sliceDef |
            sliceDef.getType().getUnderlyingType() instanceof SliceType
            and sliceDef.getAUse() = exprNode(s1).asInstruction()
            and sliceDef.getAUse() = exprNode(s2).asInstruction()
    )
}

predicate useSameSlice(IndexExpr sliceIdx, SliceExpr sliceExpr) {
    globalValueNumber(exprNode(sliceIdx.getBase())) = globalValueNumber(exprNode(sliceExpr.getBase()))
    or useSameSSASlice(sliceIdx.getBase(), sliceExpr.getBase())
    or (
        sliceIdx.getBase() instanceof SelectorExpr and sliceExpr.getBase() instanceof SelectorExpr
        and isAliasedSelector(sliceIdx.getBase().(SelectorExpr), sliceExpr.getBase().(SelectorExpr))
    )
}

/**
 * expr[i] = nil  -> sliceExprAsmt
 * resSlice = resSlice[j:]
 * 
 * type(resSlice) == []*objT
 */
from ReferenceExpr resSlice, SliceExpr sliceExpr, Assignment sliceExprAsmt, Type objT
where selfAssignSliceExpr(sliceExprAsmt, resSlice, sliceExpr)
and isCutFrontSliceExpr(sliceExpr)
and isSliceOfPtrType(resSlice.getType(), objT)
and not exists(
        Assignment nilAsmt, IndexExpr sliceIdx |
        inSameBlock(nilAsmt, sliceExprAsmt)
        and nilAssignToSliceIdx(sliceIdx, nilAsmt)
        and useSameSlice(sliceIdx, sliceExpr)
    )
select sliceExprAsmt.getLocation() as loc, isTypeSizeLarge(objT) as objTSizeLarge

// from IndexExpr sliceIdx, SliceExpr slice
// where useSameSlice(sliceIdx, slice)
// select slice


// from IndexExpr sliceIdx, SliceExpr slice
// where useSameSlice(sliceIdx, slice)
// select sliceIdx, slice

// from ExprNode s, Node d
// where s.asExpr().getType().getUnderlyingType() instanceof SliceType
//     and globalValueNumber(s) = globalValueNumber(d)
// select s, d

