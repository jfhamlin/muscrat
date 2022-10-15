package mratlang

import (
	"io"

	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/ast"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/reader"
)

type parseOptions struct {
	filename string
}

// ParseOption represents an option that can be passed to Parse.
type ParseOption func(*parseOptions)

// WithFilename sets the filename to be associated with the input.
func WithFilename(filename string) ParseOption {
	return func(o *parseOptions) {
		o.filename = filename
	}
}

func Parse(r io.RuneScanner, opts ...ParseOption) (*Program, error) {
	o := &parseOptions{}
	for _, opt := range opts {
		opt(o)
	}

	var readOpts []reader.Option
	if o.filename != "" {
		readOpts = append(readOpts, reader.WithFilename(o.filename))
	}

	rr := reader.New(r, readOpts...)
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
