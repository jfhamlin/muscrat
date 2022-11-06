package mratlang

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/ast"
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

func (env *environment) String() string {
	return fmt.Sprintf("environment:\nScope:\n%v", env.scope.printIndented("  "))
}

func (env *environment) Define(name string, value value.Value) {
	env.scope.define(name, value)
}

func (env *environment) lookup(name string) (value.Value, bool) {
	return env.scope.lookup(name)
}

func (env *environment) PushScope() value.Environment {
	wrappedEnv := *env
	newEnv := &wrappedEnv
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

type poser interface {
	Pos() ast.Pos
}

func (env *environment) errorf(n poser, format string, args ...interface{}) error {
	pos := n.Pos()
	filename := "?"
	line := "?"
	col := "?"
	if pos.Valid() {
		if pos.Filename != "" {
			filename = pos.Filename
		}
		line = fmt.Sprintf("%d", pos.Line)
		col = fmt.Sprintf("%d", pos.Column)
	}
	location := fmt.Sprintf("%s:%s:%s", filename, line, col)

	return fmt.Errorf("%s: %s", location, fmt.Sprintf(format, args...))
}

func (env *environment) Eval(n value.Value) (value.Value, error) {
	switch v := n.(type) {
	case *value.List:
		return env.evalList(v)
	case *value.Vector:
		return env.evalVector(v)
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
		case "case":
			return env.evalCase(n)
		case "and":
			return env.evalAnd(n)
		case "or":
			return env.evalOr(n)
		case "lambda":
			return env.evalLambda(n)
		case "fn":
			return env.evalFn(n)
		case "quote":
			return env.evalQuote(n)
		case "let":
			return env.evalLet(n)
		}
	}

	// otherwise, handle a function call
	var res []value.Value
	for _, item := range n.Items {
		v, err := env.Eval(item)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return env.applyFunc(res[0], res[1:])
}

func (env *environment) evalVector(n *value.Vector) (value.Value, error) {
	var res []value.Value
	for _, item := range n.Items {
		v, err := env.Eval(item)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return &value.Vector{Section: n.Section, Items: res}, nil
}

func (env *environment) evalScalar(n value.Value) (value.Value, error) {
	switch v := n.(type) {
	case *value.Symbol:
		if val, ok := env.lookup(v.Value); ok {
			return val, nil
		}
		return nil, env.errorf(n, "undefined symbol: %s", v.Value)
	default:
		// else, it's a literal
		return v, nil
	}
}

func (env *environment) applyFunc(f value.Value, args []value.Value) (value.Value, error) {
	fn, ok := f.(value.Applyer)
	if !ok {
		// TODO: the error's location should indicate the call site, not
		// the location at which the function value was defined.
		return nil, env.errorf(f, "value is not a function: %v", f)
	}
	return fn.Apply(env, args)
}

// Special forms

type nopApplyer struct{}

func (na *nopApplyer) Apply(env *environment, args []value.Value) (value.Value, error) {
	return nil, nil
}

func (env *environment) evalDef(n *value.List) (value.Value, error) {
	if len(n.Items) < 3 {
		return nil, env.errorf(n, "invalid definition, too few items")
	}

	switch v := n.Items[1].(type) {
	case *value.Symbol:
		if len(n.Items) != 3 {
			return nil, env.errorf(n, "invalid definition, too many items")
		}
		val, err := env.Eval(n.Items[2])
		if err != nil {
			return nil, err
		}
		env.Define(v.Value, val)
		return nil, nil
	case *value.List:
		if len(v.Items) == 0 {
			return nil, env.errorf(n, "invalid function definition, no name")
		}
		sym, ok := v.Items[0].(*value.Symbol)
		if !ok {
			return nil, env.errorf(n, "invalid function definition, name is not a symbol")
		}
		argNames := make([]string, 0, len(v.Items)-1)
		for _, item := range v.Items[1:] {
			argSym, ok := item.(*value.Symbol)
			if !ok {
				return nil, env.errorf(n, "invalid function definition, argument is not a symbol")
			}
			argNames = append(argNames, argSym.Value)
		}
		env.Define(sym.Value, &value.Func{
			// TODO: Section (here and elsewhere in this file) isn't quite
			// right, but close enough for now for useful errors.
			Section:    n.Section,
			LambdaName: sym.Value,
			ArgNames:   argNames,
			Exprs:      value.NewList(n.Items[2:]),
			Env:        env,
		})
		return nil, nil
	}

	return nil, env.errorf(n, "invalid definition, first item is not a symbol")
}

func (env *environment) evalLambda(n *value.List) (value.Value, error) {
	if len(n.Items) < 3 {
		return nil, env.errorf(n, "invalid lambda, need args and body")
	}
	args, ok := n.Items[1].(*value.List)
	if !ok {
		return nil, env.errorf(n, "invalid lambda, args must be a list")
	}

	argNames, err := nodeAsStringList(args)
	if err != nil {
		return nil, err
	}
	return &value.Func{
		Section:  n.Section,
		ArgNames: argNames,
		Exprs:    value.NewList(n.Items[2:]),
		Env:      env,
	}, nil
}

func (env *environment) evalFn(n *value.List) (value.Value, error) {
	if len(n.Items) < 3 {
		return nil, env.errorf(n, "invalid fn expression, need args and body")
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
		return nil, env.errorf(n, "invalid fn expression, need args and body")
	}

	args, ok := items[0].(*value.List)
	if !ok {
		return nil, env.errorf(n, "invalid fn expression, args must be a list")
	}
	argNames, err := nodeAsStringList(args)
	if err != nil {
		return nil, err
	}
	return &value.Func{
		Section:    n.Section,
		LambdaName: fnName,
		ArgNames:   argNames,
		Exprs:      value.NewList(items[1:]),
		Env:        env,
	}, nil
}

func (env *environment) evalIf(n *value.List) (value.Value, error) {
	if len(n.Items) < 3 || len(n.Items) > 4 {
		return nil, env.errorf(n, "invalid if, need `cond ifExp [elseExp]`")
	}
	cond, err := env.Eval(n.Items[1])
	if err != nil {
		return nil, err
	}

	b, ok := cond.(*value.Bool)
	if !ok || b.Value {
		res, err := env.Eval(n.Items[2])
		// non-bool is always true
		return res, err //env.Eval(n.Items[2])
	}

	if len(n.Items) == 4 {
		return env.Eval(n.Items[3])
	}
	return nil, nil
}

// cases use syntax and most of the semantics of Clojure's case (not Scheme's).
// see https://clojuredocs.org/clojure.core/case
func (env *environment) evalCase(n *value.List) (value.Value, error) {
	if len(n.Items) < 4 {
		return nil, env.errorf(n, "invalid case, need `case caseExp & caseClauses`")
	}
	cond, err := env.Eval(n.Items[1])
	if err != nil {
		return nil, err
	}

	cases := n.Items[2:]

	for len(cases) >= 2 {
		test, result := cases[0], cases[1]
		cases = cases[2:]

		testItems := []value.Value{test}
		testList, ok := test.(*value.List)
		if ok {
			testItems = testList.Items
		}

		for _, testItem := range testItems {
			if testItem.Equal(cond) {
				return env.Eval(result)
			}
		}
	}
	if len(cases) == 1 {
		return env.Eval(cases[0])
	}
	return nil, nil
}

func (env *environment) evalAnd(n *value.List) (value.Value, error) {
	if len(n.Items) < 2 {
		return nil, env.errorf(n, "invalid and, need at least one arg")
	}
	for _, item := range n.Items[1:] {
		res, err := env.Eval(item)
		if err != nil {
			return nil, err
		}
		b, ok := res.(*value.Bool)
		if !ok || !b.Value {
			if b == nil {
				return value.NewBool(false), nil
			}
			return b, nil
		}
	}
	return &value.Bool{Value: true}, nil
}

func (env *environment) evalOr(n *value.List) (value.Value, error) {
	if len(n.Items) < 2 {
		return nil, env.errorf(n, "invalid or, need at least one arg")
	}

	for _, item := range n.Items[1:] {
		res, err := env.Eval(item)
		if err != nil {
			return nil, err
		}
		b, ok := res.(*value.Bool)
		if ok && b.Value {
			return b, nil
		}
	}
	return &value.Bool{Value: false}, nil
}

func (env *environment) evalQuote(n *value.List) (value.Value, error) {
	if len(n.Items) != 2 {
		return nil, env.errorf(n, "invalid quote, need 1 argument")
	}

	return n.Items[1], nil
}

// essential syntax: let <bindings> <body>
//
// Syntax: <Bindings> should have the form
//
// ((<variable 1> <init 1>) ...),
//
// where each <init> is an expression, and <body> should be a sequence
// of one or more expressions. It is an error for a <variable> to
// appear more than once in the list of variables being bound.
//
// Semantics: The <init>s are evaluated in the current environment (in
// some unspecified order), the <variable>s are bound to fresh
// locations holding the results, the <body> is evaluated in the
// extended environment, and the value of the last expression of
// <body> is returned. Each binding of a <variable> has <body> as its
// region.
func (env *environment) evalLet(n *value.List) (value.Value, error) {
	if len(n.Items) < 3 {
		return nil, env.errorf(n, "invalid let, need bindings and body")
	}

	bindings, ok := n.Items[1].(*value.List)
	if !ok {
		return nil, env.errorf(n.Items[1], "invalid let, bindings must be a list")
	}

	// shuffle the bindings to evaluate them in a random order. this
	// prevents users from relying on the order of evaluation.
	shuffled := make([]*value.List, len(bindings.Items))
	for i, j := range rand.Perm(len(bindings.Items)) {
		item, ok := bindings.Items[i].(*value.List)
		if !ok || len(item.Items) != 2 {
			return nil, env.errorf(bindings.Items[i], "invalid let, bindings must be a list of lists of length 2")
		}
		shuffled[j] = item
	}

	// evaluate the bindings in a random order
	bindingsMap := make(map[string]value.Value)
	for _, binding := range shuffled {
		name, ok := binding.Items[0].(*value.Symbol)
		if !ok {
			return nil, env.errorf(binding.Items[0], "invalid let, binding name must be a symbol")
		}
		if _, ok := bindingsMap[name.Value]; ok {
			return nil, env.errorf(binding.Items[0], "invalid let, duplicate binding name")
		}
		val, err := env.Eval(binding.Items[1])
		if err != nil {
			return nil, err
		}
		bindingsMap[name.Value] = val
	}

	// create a new environment with the bindings
	newEnv := env.PushScope()
	for name, val := range bindingsMap {
		newEnv.Define(name, val)
	}

	// evaluate the body
	var res value.Value
	var err error
	for _, item := range n.Items[2:] {
		res, err = newEnv.Eval(item)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
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
