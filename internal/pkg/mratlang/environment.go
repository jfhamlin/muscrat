package mratlang

import (
	"context"
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
	ctx context.Context

	graph *graph.Graph
	scope *scope

	stdout io.Writer

	loadPath []string
}

func newEnvironment(ctx context.Context, stdout io.Writer) *environment {
	e := &environment{
		ctx:    ctx,
		graph:  &graph.Graph{},
		scope:  newScope(),
		stdout: stdout,
	}
	addBuiltins(e)
	return e
}

func (env *environment) Context() context.Context {
	return env.ctx
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
	var result value.Value
	var err error
	continuation := func() (value.Value, value.Continuation, error) {
		return env.eval(n)
	}

	for {
		result, continuation, err = continuation()
		if err != nil {
			return nil, err
		}
		if continuation == nil {
			return result, nil
		}
	}
}

func (env *environment) ContinuationEval(n value.Value) (value.Value, value.Continuation, error) {
	return env.eval(n)
}

func (env *environment) eval(n value.Value) (value.Value, value.Continuation, error) {
	switch v := n.(type) {
	case *value.List:
		return env.evalList(v)
	case *value.Vector:
		res, err := env.evalVector(v)
		return res, nil, err
	default:
		res, err := env.evalScalar(n)
		return res, nil, err
	}
}

func asContinuationResult(v value.Value, err error) (value.Value, value.Continuation, error) {
	return v, nil, err
}

func (env *environment) evalList(n *value.List) (value.Value, value.Continuation, error) {
	if n.IsEmpty() {
		return nil, nil, nil
	}

	first := n.Item()
	if sym, ok := first.(*value.Symbol); ok {
		// handle special forms
		switch sym.Value {
		case "def":
			return asContinuationResult(env.evalDef(n))
		case "if":
			return env.evalIf(n)
		case "case":
			return env.evalCase(n)
		case "and":
			return env.evalAnd(n)
		case "or":
			return env.evalOr(n)
		case "lambda":
			return asContinuationResult(env.evalLambda(n))
		case "fn":
			return asContinuationResult(env.evalFn(n))
		case "quote":
			return asContinuationResult(env.evalQuote(n))
		case "let":
			return env.evalLet(n)
		}
	}

	// otherwise, handle a function call
	var res []value.Value
	for cur := n; !cur.IsEmpty(); cur = cur.Next() {
		item := cur.Item()
		v, err := env.Eval(item)
		if err != nil {
			return nil, nil, err
		}
		res = append(res, v)
	}

	return nil, func() (value.Value, value.Continuation, error) {
		return env.applyFunc(res[0], res[1:])
	}, nil
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

func (env *environment) applyFunc(f value.Value, args []value.Value) (value.Value, value.Continuation, error) {
	cfn, ok := f.(value.ContinuationApplyer)
	if ok {
		return cfn.ContinuationApply(env, args)
	}

	fn, ok := f.(value.Applyer)
	if !ok {
		// TODO: the error's location should indicate the call site, not
		// the location at which the function value was defined.
		return nil, nil, env.errorf(f, "value is not a function: %v", f)
	}
	res, err := fn.Apply(env, args)
	return res, nil, err
}

// Special forms

type nopApplyer struct{}

func (na *nopApplyer) Apply(env *environment, args []value.Value) (value.Value, error) {
	return nil, nil
}

func (env *environment) evalDef(n *value.List) (value.Value, error) {
	listLength := n.Count()
	if listLength < 3 {
		return nil, env.errorf(n, "invalid definition, too few items")
	}

	switch v := n.Next().Item().(type) {
	case *value.Symbol:
		if listLength != 3 {
			return nil, env.errorf(n, "invalid definition, too many items")
		}
		val, err := env.Eval(n.Next().Next().Item())
		if err != nil {
			return nil, err
		}
		env.Define(v.Value, val)
		return nil, nil
	case *value.List:
		if v.Count() == 0 {
			return nil, env.errorf(n, "invalid function definition, no name")
		}
		sym, ok := v.Item().(*value.Symbol)
		if !ok {
			return nil, env.errorf(n, "invalid function definition, name is not a symbol")
		}
		argNames := make([]string, 0, v.Count()-1)
		for cur := v.Next(); !cur.IsEmpty(); cur = cur.Next() {
			item := cur.Item()
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
			Exprs:      n.Next().Next(),
			Env:        env,
		})
		return nil, nil
	}

	return nil, env.errorf(n, "invalid definition, first item is not a symbol")
}

func (env *environment) evalLambda(n *value.List) (value.Value, error) {
	if n.Count() < 3 {
		return nil, env.errorf(n, "invalid lambda, need args and body")
	}
	args, ok := n.Next().Item().(*value.List)
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
		Exprs:    n.Next().Next(),
		Env:      env,
	}, nil
}

func (env *environment) evalFn(n *value.List) (value.Value, error) {
	listLength := n.Count()
	if listLength < 3 {
		return nil, env.errorf(n, "invalid fn expression, need args and body")
	}

	items := make([]value.Value, 0, listLength-1)
	for cur := n.Next(); !cur.IsEmpty(); cur = cur.Next() {
		items = append(items, cur.Item())
	}

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

func (env *environment) evalIf(n *value.List) (value.Value, value.Continuation, error) {
	listLength := n.Count()
	if listLength < 3 || listLength > 4 {
		return nil, nil, env.errorf(n, "invalid if, need `cond ifExp [elseExp]`")
	}
	cond, err := env.Eval(n.Next().Item())
	if err != nil {
		return nil, nil, err
	}

	b, ok := cond.(*value.Bool)
	if !ok || b.Value { // non-bool is always true
		return nil, func() (value.Value, value.Continuation, error) {
			return env.eval(n.Next().Next().Item())
		}, nil
	}

	if listLength == 4 {
		return nil, func() (value.Value, value.Continuation, error) {
			return env.eval(n.Next().Next().Next().Item())
		}, nil
	}
	return nil, nil, nil
}

// cases use syntax and most of the semantics of Clojure's case (not Scheme's).
// see https://clojuredocs.org/clojure.core/case
func (env *environment) evalCase(n *value.List) (value.Value, value.Continuation, error) {
	listLength := n.Count()
	if listLength < 4 {
		return nil, nil, env.errorf(n, "invalid case, need `case caseExp & caseClauses`")
	}
	cond, err := env.Eval(n.Next().Item())
	if err != nil {
		return nil, nil, err
	}

	//cases := n.Items[2:]
	cases := make([]value.Value, 0, listLength-2)
	for cur := n.Next().Next(); !cur.IsEmpty(); cur = cur.Next() {
		cases = append(cases, cur.Item())
	}

	for len(cases) >= 2 {
		test, result := cases[0], cases[1]
		cases = cases[2:]

		testItems := []value.Value{test}
		testList, ok := test.(*value.List)
		if ok {
			testItems = make([]value.Value, 0, testList.Count())
			for cur := testList; !cur.IsEmpty(); cur = cur.Next() {
				testItems = append(testItems, cur.Item())
			}
		}

		for _, testItem := range testItems {
			if testItem.Equal(cond) {
				return nil, func() (value.Value, value.Continuation, error) {
					return env.eval(result)
				}, nil
			}
		}
	}
	if len(cases) == 1 {
		return nil, func() (value.Value, value.Continuation, error) {
			return env.eval(cases[0])
		}, nil
	}
	return nil, nil, nil
}

func toBool(v value.Value) *value.Bool {
	b, ok := v.(*value.Bool)
	if !ok {
		return value.NewBool(false)
	}
	return b
}

func toBoolContinuation(v value.Value, c value.Continuation, err error) (value.Value, value.Continuation, error) {
	if err != nil {
		return nil, nil, err
	}
	if c == nil {
		return toBool(v), nil, nil
	}
	return nil, func() (value.Value, value.Continuation, error) {
		return toBoolContinuation(c())
	}, nil
}

func (env *environment) evalAnd(n *value.List) (value.Value, value.Continuation, error) {
	listLength := n.Count()
	if listLength < 2 {
		return nil, nil, env.errorf(n, "invalid and, need at least one arg")
	}

	cur := n.Next()
	// iterate through all but the last item.
	// evaluate the final item in a continuation.
	for ; !cur.Next().IsEmpty(); cur = cur.Next() {
		item := cur.Item()
		res, err := env.Eval(item)
		if err != nil {
			return nil, nil, err
		}
		b, ok := res.(*value.Bool)
		if !ok || !b.Value {
			if b == nil {
				return value.NewBool(false), nil, nil
			}
			return b, nil, nil
		}
	}
	// return a continuation for the last item
	return toBoolContinuation(nil, func() (value.Value, value.Continuation, error) {
		// TODO: need to convert to bool...
		return env.eval(cur.Item())
	}, nil)
}

func (env *environment) evalOr(n *value.List) (value.Value, value.Continuation, error) {
	listLength := n.Count()
	if listLength < 2 {
		return nil, nil, env.errorf(n, "invalid or, need at least one arg")
	}

	//for _, item := range n.Items[1 : len(n.Items)-1] {
	cur := n.Next()
	// iterate through all but the last item.
	// evaluate the final item in a continuation.
	for ; !cur.Next().IsEmpty(); cur = cur.Next() {
		item := cur.Item()
		res, err := env.Eval(item)
		if err != nil {
			return nil, nil, err
		}
		b, ok := res.(*value.Bool)
		if ok && b.Value {
			return b, nil, nil
		}
	}
	// return a continuation for the last item
	return toBoolContinuation(nil, func() (value.Value, value.Continuation, error) {
		return env.eval(cur.Item())
	}, nil)
}

func (env *environment) evalQuote(n *value.List) (value.Value, error) {
	listLength := n.Count()
	if listLength != 2 {
		return nil, env.errorf(n, "invalid quote, need 1 argument")
	}

	return n.Next().Item(), nil
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
func (env *environment) evalLet(n *value.List) (value.Value, value.Continuation, error) {
	items := listAsSlice(n)
	if len(items) < 3 {
		return nil, nil, env.errorf(n, "invalid let, need bindings and body")
	}

	bindingList, ok := items[1].(*value.List)
	if !ok {
		return nil, nil, env.errorf(items[1], "invalid let, bindings must be a list")
	}
	bindings := listAsSlice(bindingList)

	// shuffle the bindings to evaluate them in a random order. this
	// prevents users from relying on the order of evaluation.
	shuffled := make([]*value.List, len(bindings))
	for i, j := range rand.Perm(len(bindings)) {
		item, ok := bindings[i].(*value.List)
		if !ok || item.Count() != 2 {
			return nil, nil, env.errorf(bindings[i], "invalid let, bindings must be a list of lists of length 2")
		}
		shuffled[j] = item
	}

	// evaluate the bindings in a random order
	bindingsMap := make(map[string]value.Value)
	for _, binding := range shuffled {
		nameValue := binding.Item()
		name, ok := nameValue.(*value.Symbol)
		if !ok {
			return nil, nil, env.errorf(nameValue, "invalid let, binding name must be a symbol")
		}
		if _, ok := bindingsMap[name.Value]; ok {
			return nil, nil, env.errorf(nameValue, "invalid let, duplicate binding name")
		}
		val, err := env.Eval(binding.Next().Item())
		if err != nil {
			return nil, nil, err
		}
		bindingsMap[name.Value] = val
	}

	// create a new environment with the bindings
	newEnv := env.PushScope().(*environment)
	for name, val := range bindingsMap {
		newEnv.Define(name, val)
	}

	// evaluate the body
	var err error
	for _, item := range items[2 : len(items)-1] {
		_, err = newEnv.Eval(item)
		if err != nil {
			return nil, nil, err
		}
	}
	// return a continuation for the last item
	return nil, func() (value.Value, value.Continuation, error) {
		return newEnv.eval(items[len(items)-1])
	}, nil
}

// Helpers

func nodeAsStringList(n *value.List) ([]string, error) {
	var res []string
	for cur := n; !cur.IsEmpty(); cur = cur.Next() {
		item := cur.Item()
		sym, ok := item.(*value.Symbol)
		if !ok {
			return nil, fmt.Errorf("invalid argument list: %v", n)
		}
		res = append(res, sym.Value)
	}
	return res, nil
}

func listAsSlice(lst *value.List) []value.Value {
	var res []value.Value
	for cur := lst; !cur.IsEmpty(); cur = cur.Next() {
		res = append(res, cur.Item())
	}
	return res
}
