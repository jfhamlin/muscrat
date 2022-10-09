package mratlang

import (
	"fmt"
	"io"

	"github.com/nsf/sexp"
)

func Parse(r io.RuneReader) (*Program, error) {
	n, err := sexp.Parse(r, nil)
	if err != nil {
		return nil, err
	}

	return newProgramFromNode(n)
}

func newProgramFromNode(n *sexp.Node) (*Program, error) {
	p := &Program{
		node: n,
	}

	return p, nil
}

func printAST(n *sexp.Node, indent int) {
	for i := 0; i < indent; i++ {
		fmt.Print(" ")
	}
	if n.IsList() {
		fmt.Printf("(%s\n", n.Value)
	} else {
		fmt.Println(n.Value)
	}
	child := n.Children
	for child != nil {
		printAST(child, indent+1)
		child = child.Next
	}
}
