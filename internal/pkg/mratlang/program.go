package mratlang

import (
	"context"
	"io"
	"os"

	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/value"
)

type Program struct {
	nodes []value.Value
}

type evalOptions struct {
	stdout   io.Writer
	loadPath []string
	env      *environment
}

type EvalOption func(*evalOptions)

func WithStdout(w io.Writer) EvalOption {
	return func(opts *evalOptions) {
		opts.stdout = w
	}
}

func WithLoadPath(path []string) EvalOption {
	return func(opts *evalOptions) {
		opts.loadPath = path
	}
}

func withEnv(env value.Environment) EvalOption {
	e := env.(*environment)
	return func(opts *evalOptions) {
		opts.env = e
	}
}

func (p *Program) Eval(opts ...EvalOption) (*graph.Graph, []graph.SinkChan, error) {
	options := &evalOptions{
		stdout: os.Stdout,
	}
	for _, opt := range opts {
		opt(options)
	}

	env := options.env
	if env == nil {
		env = newEnvironment(context.Background(), options.stdout)
		env.loadPath = options.loadPath
	}

	for _, node := range p.nodes {
		_, err := env.Eval(node)
		if err != nil {
			return nil, nil, err
		}
	}

	var sinkChans []graph.SinkChan
	for _, sink := range env.Graph().Sinks() {
		sinkChans = append(sinkChans, sink.Chan())
	}

	return env.graph, sinkChans, nil
}
