package value

import (
	"fmt"

	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/ast"
)

// FromAST converts an AST node to a value.
func FromAST(node ast.Node) Value {
	switch node := node.(type) {
	case *ast.Number:
		return NewNum(node.Value, WithSection(node.Section))
	case *ast.String:
		return NewStr(node.Value, WithSection(node.Section))
	case *ast.Bool:
		return NewBool(node.Value, WithSection(node.Section))
	case *ast.Keyword:
		return NewKeyword(node.Value, WithSection(node.Section))
	case *ast.List:
		var items []Value
		for _, item := range node.Items {
			items = append(items, FromAST(item))
		}
		return NewList(items, WithSection(node.Section))
	case *ast.Vector:
		var items []Value
		for _, item := range node.Items {
			items = append(items, FromAST(item))
		}
		return NewVector(items, WithSection(node.Section))
	case *ast.Symbol:
		return NewSymbol(node.Value, WithSection(node.Section))
	default:
		panic(fmt.Sprintf("unhandled node type: %T", node))
	}
	return nil
}
