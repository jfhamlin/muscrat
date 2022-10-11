package mratlang

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/ast"
)

type Program struct {
	nodes []ast.Node
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

func withEnv(env *environment) EvalOption {
	return func(opts *evalOptions) {
		opts.env = env
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
		env = newEnvironment(options.stdout)
		env.loadPath = options.loadPath
	}

	for _, node := range p.nodes {
		_, err := env.evalNode(node)
		if err != nil {
			return nil, nil, err
		}
	}

	// TODO: sink info should just be part of the graph.
	return env.graph, env.sinks.sinkChannels, nil
}

type sinks struct {
	sinkNodeIDs  []graph.NodeID
	sinkChannels []graph.SinkChan
}

type environment struct {
	graph *graph.Graph
	sinks *sinks
	scope *scope

	stdout io.Writer

	loadPath []string
}

func newEnvironment(stdout io.Writer) *environment {
	e := &environment{
		graph:  &graph.Graph{},
		sinks:  &sinks{},
		scope:  newScope(),
		stdout: stdout,
	}
	addBuiltins(e)
	return e
}

func (env *environment) define(name string, value Value) {
	env.scope.define(name, value)
}

func (env *environment) lookup(name string) (Value, bool) {
	return env.scope.lookup(name)
}

func (env *environment) pushScope() *environment {
	newEnv := &(*env)
	newEnv.scope = newEnv.scope.push()
	return newEnv
}

func (env *environment) pushLoadPaths(paths []string) *environment {
	newEnv := &(*env)
	newEnv.loadPath = append(paths, newEnv.loadPath...)
	return newEnv
}

func (env *environment) resolveFile(filename string) (string, bool) {
	if filepath.IsAbs(filename) {
		return filename, true
	}

	for _, path := range env.loadPath {
		fullPath := filepath.Join(path, filename)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, true
		}
	}
	return "", false
}

func (env *environment) evalNode(n ast.Node) (Value, error) {
	switch v := n.(type) {
	case *ast.List:
		return env.evalList(v)
	default:
		return env.evalScalar(n)
	}
}

func (env *environment) evalList(n *ast.List) (Value, error) {
	if len(n.Items) == 0 {
		return nil, nil
	}

	first := n.Items[0]
	if sym, ok := first.(*ast.Symbol); ok {
		// handle special forms
		switch sym.Value {
		case "def":
			return env.evalDef(n)
		case "if":
			return env.evalIf(n)
		case "lambda":
			return env.evalLambda(n)
		case "fn":
			return env.evalFn(n)
		}
	}

	// otherwise, handle a function call
	var res []Value
	for _, item := range n.Items {
		v, err := env.evalNode(item)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return env.applyFunc(res[0], res[1:])
}

func (env *environment) evalScalar(n ast.Node) (Value, error) {
	switch v := n.(type) {
	case *ast.Number:
		return &Num{v.Value}, nil
	case *ast.Quote:
		panic("unimplemented")
	case *ast.Symbol:
		if val, ok := env.lookup(v.Value); ok {
			return val, nil
		}
		fmt.Println("unbound symbol:", n)
		return nil, fmt.Errorf("XXX undefined symbol: %s", v.Value)
	case *ast.Keyword:
		return &Keyword{Value: v.Value}, nil
	case *ast.String:
		return &Str{Value: v.Value}, nil
	case *ast.Bool:
		return &Bool{Value: v.Value}, nil
	default:
		return nil, fmt.Errorf("unhandled scalar type: %T", n)
	}
}

func (env *environment) applyFunc(f Value, args []Value) (Value, error) {
	fn, ok := f.(applyer)
	if !ok {
		return nil, fmt.Errorf("not a function: %v", f)
	}
	return fn.Apply(env.pushScope(), args)
}

// Special forms

type applyer interface {
	Apply(env *environment, args []Value) (Value, error)
}

type applyerFunc func(env *environment, args []Value) (Value, error)

func (f applyerFunc) Apply(env *environment, args []Value) (Value, error) {
	return f(env, args)
}

type nopApplyer struct{}

func (na *nopApplyer) Apply(env *environment, args []Value) (Value, error) {
	return nil, nil
}

func (env *environment) evalDef(n *ast.List) (Value, error) {
	if len(n.Items) < 3 {
		return nil, fmt.Errorf("invalid def: %v", n.Pos())
	}

	switch v := n.Items[1].(type) {
	case *ast.Symbol:
		if len(n.Items) != 3 {
			return nil, fmt.Errorf("invalid def: %v", n.Pos())
		}
		val, err := env.evalNode(n.Items[2])
		if err != nil {
			return nil, err
		}
		env.define(v.Value, val)
		return nil, nil
	case *ast.List:
		if len(v.Items) == 0 {
			return nil, fmt.Errorf("invalid def: %v", n.Pos())
		}
		sym, ok := v.Items[0].(*ast.Symbol)
		if !ok {
			return nil, fmt.Errorf("invalid def: %v", n.Pos())
		}
		argNames := make([]string, 0, len(v.Items)-1)
		for _, item := range v.Items[1:] {
			argSym, ok := item.(*ast.Symbol)
			if !ok {
				return nil, fmt.Errorf("invalid def: %v", n.Pos())
			}
			argNames = append(argNames, argSym.Value)
		}
		env.define(sym.Value, &Func{
			argNames: argNames,
			node:     ast.NewList(n.Items[2:], ast.Section{StartPos: n.Pos(), EndPos: n.End()}),
			env:      env,
		})
		return nil, nil
	}

	return nil, fmt.Errorf("invalid def: %v", n.Pos())
}

func (env *environment) evalLambda(n *ast.List) (Value, error) {
	if len(n.Items) < 3 {
		return nil, fmt.Errorf("invalid lambda, need args and body: %v", n)
	}
	args, ok := n.Items[1].(*ast.List)
	if !ok {
		return nil, fmt.Errorf("invalid lambda, args must be a list: %v", n)
	}

	argNames, err := nodeAsStringList(args)
	if err != nil {
		return nil, err
	}
	return &Func{
		argNames: argNames,
		node:     ast.NewList(n.Items[2:], ast.Section{StartPos: n.Pos(), EndPos: n.End()}),
		env:      env,
	}, nil
}

func (env *environment) evalFn(n *ast.List) (Value, error) {
	if len(n.Items) < 3 {
		return nil, fmt.Errorf("invalid fn expression, need args and body: %v", n)
	}

	items := n.Items[1:]

	var fnName string
	if sym, ok := items[0].(*ast.Symbol); ok {
		// if the first child is not a list, it's the name of the
		// function. this can be used for recursion.
		fnName = sym.Value
		items = items[1:]
	}

	if len(items) < 2 {
		return nil, fmt.Errorf("invalid fn expression, need args and body: %v", n)
	}

	args, ok := items[0].(*ast.List)
	if !ok {
		return nil, fmt.Errorf("invalid fn expression, args must be a list: %v", n)
	}
	argNames, err := nodeAsStringList(args)
	if err != nil {
		return nil, err
	}
	return &Func{
		lambdaName: fnName,
		argNames:   argNames,
		node:       ast.NewList(items[1:], ast.Section{StartPos: n.Pos(), EndPos: n.End()}),
		env:        env,
	}, nil
}

func (env *environment) evalIf(n *ast.List) (Value, error) {
	if len(n.Items) < 3 || len(n.Items) > 4 {
		return nil, fmt.Errorf("invalid if, need `cond ifExp [elseExp]`: %v", n)
	}
	cond, err := env.evalNode(n.Items[1])
	if err != nil {
		return nil, err
	}

	b, ok := cond.(*Bool)
	if !ok || b.Value {
		// non-bool is always true
		return env.evalNode(n.Items[2])
	}
	if len(n.Items) == 4 {
		return env.evalNode(n.Items[3])
	}
	return nil, nil
}

// Helpers

func nodeAsStringList(n *ast.List) ([]string, error) {
	var res []string
	for _, item := range n.Items {
		sym, ok := item.(*ast.Symbol)
		if !ok {
			return nil, fmt.Errorf("invalid argument list: %v", n)
		}
		res = append(res, sym.Value)
	}
	return res, nil
}
