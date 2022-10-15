package mratlang

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/value"
)

type environment struct {
	graph *graph.Graph
	scope *scope

	stdout io.Writer

	loadPath []string
}

func newEnvironment(stdout io.Writer) *environment {
	e := &environment{
		graph:  &graph.Graph{},
		scope:  newScope(),
		stdout: stdout,
	}
	addBuiltins(e)
	return e
}

func (env *environment) Define(name string, value value.Value) {
	env.scope.define(name, value)
}

func (env *environment) lookup(name string) (value.Value, bool) {
	return env.scope.lookup(name)
}

func (env *environment) PushScope() value.Environment {
	newEnv := &(*env)
	newEnv.scope = newEnv.scope.push()
	return newEnv
}

func (env *environment) Graph() *graph.Graph {
	return env.graph
}

func (env *environment) Stdout() io.Writer {
	return env.stdout
}

func (env *environment) PushLoadPaths(paths []string) value.Environment {
	newEnv := &(*env)
	newEnv.loadPath = append(paths, newEnv.loadPath...)
	return newEnv
}

func (env *environment) ResolveFile(filename string) (string, bool) {
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

func (env *environment) Eval(v value.Value) (value.Value, error) {
	return env.evalNode(v)
}

func (env *environment) evalNode(n value.Value) (value.Value, error) {
	switch v := n.(type) {
	case *value.List:
		return env.evalList(v)
	default:
		return env.evalScalar(n)
	}
}

func (env *environment) evalList(n *value.List) (value.Value, error) {
	if len(n.Items) == 0 {
		return nil, nil
	}

	first := n.Items[0]
	if sym, ok := first.(*value.Symbol); ok {
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
	var res []value.Value
	for _, item := range n.Items {
		v, err := env.evalNode(item)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return env.applyFunc(res[0], res[1:])
}

func (env *environment) evalScalar(n value.Value) (value.Value, error) {
	switch v := n.(type) {
	case *value.Num:
		return &value.Num{v.Value}, nil
	case *value.Symbol:
		if val, ok := env.lookup(v.Value); ok {
			return val, nil
		}
		fmt.Println("unbound symbol:", n)
		return nil, fmt.Errorf("XXX undefined symbol: %s", v.Value)
	case *value.Keyword:
		return &value.Keyword{Value: v.Value}, nil
	case *value.Str:
		return &value.Str{Value: v.Value}, nil
	case *value.Bool:
		return &value.Bool{Value: v.Value}, nil
	default:
		return nil, fmt.Errorf("unhandled scalar type: %T", n)
	}
}

func (env *environment) applyFunc(f value.Value, args []value.Value) (value.Value, error) {
	fn, ok := f.(value.Applyer)
	if !ok {
		return nil, fmt.Errorf("not a function: %v", f)
	}
	return fn.Apply(env.PushScope(), args)
}

// Special forms

type nopApplyer struct{}

func (na *nopApplyer) Apply(env *environment, args []value.Value) (value.Value, error) {
	return nil, nil
}

func (env *environment) evalDef(n *value.List) (value.Value, error) {
	if len(n.Items) < 3 {
		return nil, fmt.Errorf("invalid def: %v", n.Pos())
	}

	switch v := n.Items[1].(type) {
	case *value.Symbol:
		if len(n.Items) != 3 {
			return nil, fmt.Errorf("invalid def: %v", n.Pos())
		}
		val, err := env.evalNode(n.Items[2])
		if err != nil {
			return nil, err
		}
		env.Define(v.Value, val)
		return nil, nil
	case *value.List:
		if len(v.Items) == 0 {
			return nil, fmt.Errorf("invalid def: %v", n.Pos())
		}
		sym, ok := v.Items[0].(*value.Symbol)
		if !ok {
			return nil, fmt.Errorf("invalid def: %v", n.Pos())
		}
		argNames := make([]string, 0, len(v.Items)-1)
		for _, item := range v.Items[1:] {
			argSym, ok := item.(*value.Symbol)
			if !ok {
				return nil, fmt.Errorf("invalid def: %v", n.Pos())
			}
			argNames = append(argNames, argSym.Value)
		}
		env.Define(sym.Value, &value.Func{
			ArgNames: argNames,
			Exprs:    value.NewList(n.Items[2:]),
			Env:      env,
		})
		return nil, nil
	}

	return nil, fmt.Errorf("invalid def: %v", n.Pos())
}

func (env *environment) evalLambda(n *value.List) (value.Value, error) {
	if len(n.Items) < 3 {
		return nil, fmt.Errorf("invalid lambda, need args and body: %v", n)
	}
	args, ok := n.Items[1].(*value.List)
	if !ok {
		return nil, fmt.Errorf("invalid lambda, args must be a list: %v", n)
	}

	argNames, err := nodeAsStringList(args)
	if err != nil {
		return nil, err
	}
	return &value.Func{
		ArgNames: argNames,
		Exprs:    value.NewList(n.Items[2:]),
		Env:      env,
	}, nil
}

func (env *environment) evalFn(n *value.List) (value.Value, error) {
	if len(n.Items) < 3 {
		return nil, fmt.Errorf("invalid fn expression, need args and body: %v", n)
	}

	items := n.Items[1:]

	var fnName string
	if sym, ok := items[0].(*value.Symbol); ok {
		// if the first child is not a list, it's the name of the
		// function. this can be used for recursion.
		fnName = sym.Value
		items = items[1:]
	}

	if len(items) < 2 {
		return nil, fmt.Errorf("invalid fn expression, need args and body: %v", n)
	}

	args, ok := items[0].(*value.List)
	if !ok {
		return nil, fmt.Errorf("invalid fn expression, args must be a list: %v", n)
	}
	argNames, err := nodeAsStringList(args)
	if err != nil {
		return nil, err
	}
	return &value.Func{
		LambdaName: fnName,
		ArgNames:   argNames,
		Exprs:      value.NewList(items[1:]),
		Env:        env,
	}, nil
}

func (env *environment) evalIf(n *value.List) (value.Value, error) {
	if len(n.Items) < 3 || len(n.Items) > 4 {
		return nil, fmt.Errorf("invalid if, need `cond ifExp [elseExp]`: %v", n)
	}
	cond, err := env.evalNode(n.Items[1])
	if err != nil {
		return nil, err
	}

	b, ok := cond.(*value.Bool)
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

func nodeAsStringList(n *value.List) ([]string, error) {
	var res []string
	for _, item := range n.Items {
		sym, ok := item.(*value.Symbol)
		if !ok {
			return nil, fmt.Errorf("invalid argument list: %v", n)
		}
		res = append(res, sym.Value)
	}
	return res, nil
}
