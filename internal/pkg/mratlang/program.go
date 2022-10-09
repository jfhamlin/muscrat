package mratlang

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"unicode"

	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/nsf/sexp"
)

type Program struct {
	node *sexp.Node
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

	for cur := p.node.Children; cur != nil; cur = cur.Next {
		_, err := env.evalNode(cur)
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

func (env *environment) evalNode(n *sexp.Node) (Value, error) {
	if nodeIsList(n) {
		return env.evalList(n)
	}

	return env.evalScalar(n)
}

func (env *environment) evalList(n *sexp.Node) (Value, error) {
	if n.NumChildren() == 0 {
		return nil, nil
	}

	first := n.Children

	// handle special forms
	switch first.Value {
	case "def":
		return env.evalDef(first.Next)
	case "if":
		return env.evalIf(first.Next)
	case "lambda":
		return env.evalLambda(first.Next)
	case "fn":
		return env.evalFn(first.Next)
	}

	// otherwise, handle a function call
	var res []Value
	for cur := first; cur != nil; cur = cur.Next {
		v, err := env.evalNode(cur)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return env.applyFunc(res[0], res[1:])
}

func (env *environment) evalScalar(n *sexp.Node) (Value, error) {
	valueRunes := []rune(n.Value)
	if len(valueRunes) == 0 {
		fmt.Println("empty scalar", n.Value)
	}
	firstRune := valueRunes[0]
	if unicode.IsDigit(firstRune) || firstRune == '-' || firstRune == '.' {
		f, err := strconv.ParseFloat(n.Value, 64)
		if err != nil {
			return nil, err
		}
		return &Num{Value: f}, nil
	}

	// The output of the s-expression parser being used can't
	// disambiguate between identifiers and double-quoted strings. For
	// now we just support strings with no spaces, prefixed by a single
	// quote.
	if firstRune == '\'' {
		return &Str{Value: n.Value[1:]}, nil
	}

	switch n.Value {
	case "#t":
		return &Bool{Value: true}, nil
	case "#f":
		return &Bool{Value: false}, nil
	}

	// else, it's a symbol
	if v, ok := env.lookup(n.Value); ok {
		return v, nil
	}
	return nil, fmt.Errorf("undefined symbol: %s", n.Value)
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

func (env *environment) evalDef(n *sexp.Node) (Value, error) {
	if nodeLen(n) != 2 {
		return nil, fmt.Errorf("invalid def: %v", n.Location)
	}
	if nodeIsList(n) {
		argNames, err := nodeAsStringList(n.Children.Next)
		if err != nil {
			return nil, err
		}
		env.define(n.Children.Value, &Func{
			argNames: argNames,
			node:     n.Next,
			env:      env,
		})
		return nil, nil
	}

	name := n.Value
	val, err := env.evalNode(n.Next)
	if err != nil {
		return nil, err
	}
	env.define(name, val)
	return nil, nil
}

func (env *environment) evalLambda(n *sexp.Node) (Value, error) {
	if nodeLen(n) != 2 {
		return nil, fmt.Errorf("invalid lambda, need args and body: %v", n.Location)
	}
	if !nodeIsList(n) {
		return nil, fmt.Errorf("invalid lambda, args must be a list: %v", n.Location)
	}
	argNames, err := nodeAsStringList(n.Children)
	if err != nil {
		return nil, err
	}
	return &Func{
		argNames: argNames,
		node:     n.Next,
		env:      env,
	}, nil
}

func (env *environment) evalFn(n *sexp.Node) (Value, error) {
	if nodeLen(n) < 2 {
		return nil, fmt.Errorf("invalid fn expression, need args and body: %v", n.Location)
	}

	var fnName string
	if !nodeIsList(n) {
		// if the first child is not a list, it's the name of the
		// function. this can be used for recursion.
		fnName = n.Value
		n = n.Next
	}

	if nodeLen(n) < 2 {
		return nil, fmt.Errorf("invalid fn expression, need args and body: %v", n.Location)
	}

	if !nodeIsList(n) {
		return nil, fmt.Errorf("invalid fn expression, args must be a list: %v", n.Location)
	}
	argNames, err := nodeAsStringList(n.Children)
	if err != nil {
		return nil, err
	}
	return &Func{
		lambdaName: fnName,
		argNames:   argNames,
		node:       n.Next,
		env:        env,
	}, nil
}

func (env *environment) evalIf(n *sexp.Node) (Value, error) {
	if nodeLen(n) < 2 || nodeLen(n) > 3 {
		return nil, fmt.Errorf("invalid if, need `cond ifExp [elseExp]`: %v", n.Location)
	}
	cond, err := env.evalNode(n)
	if err != nil {
		return nil, err
	}

	b, ok := cond.(*Bool)
	if !ok || b.Value {
		// non-bool is always true
		return env.evalNode(n.Next)
	}
	if n.Next.Next != nil {
		return env.evalNode(n.Next.Next)
	}
	return nil, nil
}

// Helpers

func nodeLen(n *sexp.Node) int {
	var i int
	for cur := n; cur != nil; cur = cur.Next {
		i++
	}
	return i
}

func nodeAsStringList(n *sexp.Node) ([]string, error) {
	var res []string
	for cur := n; cur != nil; cur = cur.Next {
		if cur.IsList() {
			return nil, fmt.Errorf("invalid argument list: %v", n.Location)
		}
		res = append(res, cur.Value)
	}
	return res, nil
}

func nodeIsList(n *sexp.Node) bool {
	return n.Children != nil || n.Value == ""
}
