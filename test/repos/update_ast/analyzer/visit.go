package analyzer

import "go/ast"

type unit struct{}

type abstractVisitor[Result any] interface {
	visit(node ast.Node)Result
}

func walkNodeList[ResultType any, ListType ast.Node, Visitor abstractVisitor[ResultType]](v Visitor, list []ListType) []ResultType {
	result := make([]ResultType,len(list))
	for i, x := range list {
		result[i] = v.visit(x)
	}
	return result
}
