package value

import (
	"fmt"

	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/ast"
)

// FromAST converts an AST node to a value.
func FromAST(node ast.Node) Value {
	switch node := node.(type) {
	case *ast.Number:
		return NewNum(node.Value)
	case *ast.String:
		return NewStr(node.Value)
	case *ast.Bool:
		return NewBool(node.Value)
	case *ast.Keyword:
		return NewKeyword(node.Value)
	case *ast.List:
		var items []Value
		for _, item := range node.Items {
			items = append(items, FromAST(item))
		}
		return NewList(items)
	case *ast.Symbol:
		return NewSymbol(node.Value)
	default:
		panic(fmt.Sprintf("unhandled node type: %T", node))
	}
	return nil
}
