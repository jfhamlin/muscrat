package mratlang

import (
	"io"

	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/ast"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/reader"
)

func Parse(r io.RuneScanner) (*Program, error) {
	rr := reader.New(r)
	nodes, err := rr.ReadAll()
	if err != nil {
		return nil, err
	}

	return newProgramFromNode(nodes)
}

func newProgramFromNode(nodes []ast.Node) (*Program, error) {
	p := &Program{
		nodes: nodes,
	}

	return p, nil
}
