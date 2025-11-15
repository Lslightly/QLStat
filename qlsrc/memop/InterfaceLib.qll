/**
 * Interface related Library
 */
import go

// true if allocate when converting to interface else false
predicate allocWhenConvToIface(Type type) {
    not (
           type.getUnderlyingType() instanceof InterfaceType
        or type.getUnderlyingType() instanceof PointerType
    )
}

string convSrcTypeSummary(Type type) {
    if allocWhenConvToIface(type) then
        result = "alloc"
    else if type.getUnderlyingType() instanceof InterfaceType then
        result = "interface"
    else
        result = "pointer"
}