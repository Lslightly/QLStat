/**
 * 
 * Type Helper Library
 * 
 */
import go

/**
 * return the level of pointer type for `type` through `depth`, removing type alias definition according to `noTypeAlias`
 */
predicate pointerDepth(Type type, Boolean noTypeAlias, int depth) {
    if noTypeAlias = true then
        if not (type.getUnderlyingType() instanceof PointerType) then
            depth = 0
        else
            exists(int subDepth | pointerDepth(type.getUnderlyingType().(PointerType).getBaseType(), true, subDepth) | depth = 1+subDepth)
    else
        if (not type instanceof PointerType) then
            depth = 0
        else
            exists(int subDepth | pointerDepth(type.(PointerType).getBaseType(), false, subDepth) | depth=1+subDepth)
}

/**
 * return the level of pointer type for `type`, removing type alias definition according to `noTypeAlias`
 */
int pointerDepthFn(Type type, Boolean noTypeAlias) {
    pointerDepth(type, noTypeAlias, result)
}

/**
 * type size, i.e. the size of allocation of objects
 * -1 means unknown
 */
language[monotonicAggregates]
int typeSize(Type type) {
    exists(Type t | type.getUnderlyingType() = t |
        if t instanceof BoolType then
            result = 8
        else if t instanceof IntegerType then
            result = 8
        else if t instanceof UintptrType then
            result = 8
        else if t instanceof NumericType then
            result = t.(NumericType).getASize() / 8
        else if t instanceof StringType then
            /*
            * struct {
            *  header  unsafe.Pointer
            *  len     unsafe.Pointer
            * }
            */
            result = 16
        else if t instanceof ArrayType then
            result = forwardUnknown(t.(ArrayType).getLength() * typeSize(t.(ArrayType).getElementType()), typeSize(t.(ArrayType).getElementType()))
        else if t instanceof SliceType then
            /*
            * struct {
            *  header  unsafe.Pointer
            *  len     unsafe.Pointer
            *  cap     unsafe.Pointer
            * }
            */
            result = 24
        else if t instanceof StructType then
            if exists(Field f| f = t.getField(_) and -1 = typeSize(f.getType())) then
                result = -1
            else
                result = sum(Field f| f = t.getField(_)|typeSize(f.getType()))
        else if t instanceof PointerType then
            result = 8
        else if t instanceof SignatureType then
            result = getASizeForSigatureType(t.(SignatureType))
        else if t instanceof InterfaceType then
            /*
             * type descriptor
             * data
             */
            result = 16
        else
            result = -1 // -1 means unknown
    )
}

language[monotonicAggregates]
Boolean isTypeSizeKnown(Type type) {
    exists(Type t | type.getUnderlyingType() = t |
        if t instanceof BoolType then
            result = true
        else if t instanceof IntegerType then
            result = true
        else if t instanceof UintptrType then
            result = true
        else if t instanceof NumericType then
            result = true
        else if t instanceof StringType then
            /*
            * struct {
            *  header  unsafe.Pointer
            *  len     unsafe.Pointer
            * }
            */
            result = true
        else if t instanceof ArrayType then
            result = true
        else if t instanceof SliceType then
            /*
            * struct {
            *  header  unsafe.Pointer
            *  len     unsafe.Pointer
            *  cap     unsafe.Pointer
            * }
            */
            result = true
        else if t instanceof StructType then
            if exists(string i| isTypeSizeKnown(t.getField(i).getType()) = false and t.getField(i).getType() != type) then
                result = false
            else
                result = true
        else if t instanceof PointerType then
            result = true
        else if t instanceof SignatureType then
            result = true
        else if t instanceof InterfaceType then
            /*
             * type descriptor
             * data
             */
            result = true
        else
            result = false
    )
}

bindingset[ok, err]
int forwardUnknown(int ok, int err) {
    if err = -1 then
        result = -1
    else
        result = ok
}

/**
 * 
 * get size for function type
 * 
 */
int getASizeForSigatureType(SignatureType type) {
    /*
     * struct {
     *      fn     unsafe.Pointer
     *      arg    unsafe.Pointer
     * }
     */
    result = 16
}