package mratlang

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/value"
)

type environment struct {
	ctx context.Context

	graph *graph.Graph
	scope *scope

	macros map[string]*value.Func

	gensymCounter int

	stdout io.Writer

	loadPath []string
}

func newEnvironment(ctx context.Context, stdout io.Writer) *environment {
	e := &environment{
		ctx:    ctx,
		graph:  &graph.Graph{},
		scope:  newScope(),
		macros: make(map[string]*value.Func),
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
	// TODO: define should define globally!
	env.scope.define(name, value)
}

func (env *environment) DefineMacro(name string, fn *value.Func) {
	env.macros[name] = fn
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
	Pos() value.Pos
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
		case "fn":
			return asContinuationResult(env.evalFn(n))
		case "quote":
			return asContinuationResult(env.evalQuote(n))
		case "quasiquote":
			return asContinuationResult(env.evalQuasiquote(n))
		case "let":
			return env.evalLet(n)
		case "defmacro":
			return asContinuationResult(env.evalDefMacro(n))
		}

		// handle macros
		if macro, ok := env.macros[sym.Value]; ok {
			return env.applyMacro(macro, n.Next())
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
	for i := 0; i < n.Count(); i++ {
		item := n.ValueAt(i)
		v, err := env.Eval(item)
		if err != nil {
			return nil, err
		}
		res = append(res, v)
	}
	return value.NewVector(res, value.WithSection(n.Section)), nil
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
	case *value.List: // scheme-style definition
		if v.Count() == 0 {
			return nil, env.errorf(n, "invalid function definition, no name")
		}
		sym, ok := v.Item().(*value.Symbol)
		if !ok {
			return nil, env.errorf(n, "invalid function definition, name is not a symbol: %v (%T)", v.Item(), v.Item())
		}
		var args []value.Value
		for cur := v.Next(); !cur.IsEmpty(); cur = cur.Next() {
			item := cur.Item()
			args = append(args, item)
		}
		env.Define(sym.Value, &value.Func{
			// TODO: Section (here and elsewhere in this file) isn't quite
			// right, but close enough for now for useful errors.
			Section:    n.Section,
			LambdaName: sym.Value,
			Env:        env,
			Arities: []value.FuncArity{
				{
					BindingForm: value.NewVector(args, value.WithSection(n.Section)),
					Exprs:       n.Next().Next(),
				},
			},
		})
		return nil, nil
	}

	return nil, env.errorf(n, "invalid definition, first item is not a symbol")
}

func (env *environment) evalFn(n *value.List) (value.Value, error) {
	listLength := n.Count()
	items := make([]value.Value, 0, listLength-1)
	for cur := n.Next(); !cur.IsEmpty(); cur = cur.Next() {
		items = append(items, cur.Item())
	}

	if len(items) < 2 {
		return nil, env.errorf(n, "invalid fn expression, need args and body")
	}

	var fnName string
	if sym, ok := items[0].(*value.Symbol); ok {
		// if the first child is not a list, it's the name of the
		// function. this can be used for recursion.
		fnName = sym.Value
		items = items[1:]
	}

	if len(items) == 0 {
		return nil, env.errorf(n, "invalid fn expression, need args and body")
	}

	const errorString = "invalid fn expression, expected (fn ([bindings0] body0) ([bindings1] body1) ...) or (fn [bindings] body)"

	arities := make([]*value.List, 0, len(items))
	if _, ok := items[0].(*value.Vector); ok {
		// if the next child is a vector, it's the bindings, and we only
		// have one arity.
		arities = append(arities, value.NewList(items, value.WithSection(n.Section)))
	} else {
		// otherwise, every remaining child must be a list of function
		// bindings and bodies for each arity.
		for _, item := range items {
			list, ok := item.(*value.List)
			if !ok {
				return nil, env.errorf(n, errorString)
			}
			arities = append(arities, list)
		}
	}

	arityValues := make([]value.FuncArity, len(arities))
	for i, arity := range arities {
		bindings, ok := arity.Item().(*value.Vector)
		if !ok {
			return nil, env.errorf(n, errorString)
		}
		if !value.IsValidBinding(bindings) {
			return nil, env.errorf(n, "invalid fn expression, invalid binding (%v). Must be valid destructure form", bindings)
		}

		body := arity.Next()

		arityValues[i] = value.FuncArity{
			BindingForm: bindings,
			Exprs:       body,
		}
	}
	return &value.Func{
		Section:    n.Section,
		LambdaName: fnName,
		Env:        env,
		Arities:    arityValues,
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

	if value.IsTruthy(cond) {
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

func (env *environment) evalQuote(n *value.List) (value.Value, error) {
	listLength := n.Count()
	if listLength != 2 {
		return nil, env.errorf(n, "invalid quote, need 1 argument")
	}

	return n.Next().Item(), nil
}

func (env *environment) evalQuasiquote(n *value.List) (value.Value, error) {
	listLength := n.Count()
	if listLength != 2 {
		return nil, env.errorf(n, "invalid quasiquote, need 1 argument")
	}

	// symbolNameMap tracks the names of symbols that have been renamed.
	// symbols that end with a '#' have '#' replaced with a unique
	// suffix.
	symbolNameMap := make(map[string]string)
	return env.evalQuasiquoteItem(symbolNameMap, n.Next().Item())
}

func (env *environment) evalQuasiquoteItem(symbolNameMap map[string]string, item value.Value) (value.Value, error) {
	switch item := item.(type) {
	case *value.List:
		if item.IsEmpty() {
			return item, nil
		}
		if item.Item().Equal(value.SymbolUnquote) {
			return env.Eval(item.Next().Item())
		}
		if item.Item().Equal(value.SymbolSpliceUnquote) {
			return nil, env.errorf(item, "splice-unquote not in list")
		}

		var resultValues []value.Value
		for cur := item; !cur.IsEmpty(); cur = cur.Next() {
			if lst, ok := cur.Item().(*value.List); ok && !lst.IsEmpty() && lst.Item().Equal(value.SymbolSpliceUnquote) {
				res, err := env.Eval(lst.Next().Item())
				if err != nil {
					return nil, err
				}
				vals, ok := res.(value.Nther)
				if !ok {
					return nil, env.errorf(lst, "splice-unquote did not return an enumerable")
				}
				for i := 0; ; i++ {
					v, ok := vals.Nth(i)
					if !ok {
						break
					}
					resultValues = append(resultValues, v)
				}
				continue
			}

			result, err := env.evalQuasiquoteItem(symbolNameMap, cur.Item())
			if err != nil {
				return nil, err
			}
			resultValues = append(resultValues, result)
		}
		return value.NewList(resultValues), nil
	case *value.Vector:
		if item.Count() == 0 {
			return item, nil
		}

		var resultValues []value.Value
		for i := 0; i < item.Count(); i++ {
			cur := item.ValueAt(i)
			if lst, ok := cur.(*value.List); ok && !lst.IsEmpty() && lst.Item().Equal(value.SymbolSpliceUnquote) {
				res, err := env.Eval(lst.Next().Item())
				if err != nil {
					return nil, err
				}
				vals, ok := res.(value.Nther)
				if !ok {
					return nil, env.errorf(lst, "splice-unquote did not return an enumerable")
				}
				for j := 0; ; j++ {
					v, ok := vals.Nth(j)
					if !ok {
						break
					}
					resultValues = append(resultValues, v)
				}
				continue
			}

			result, err := env.evalQuasiquoteItem(symbolNameMap, cur)
			if err != nil {
				return nil, err
			}
			resultValues = append(resultValues, result)
		}
		return value.NewVector(resultValues), nil
	case *value.Symbol:
		if !strings.HasSuffix(item.Value, "#") {
			return item, nil
		}
		newName, ok := symbolNameMap[item.Value]
		if !ok {
			newName = item.Value[:len(item.Value)-1] + "__" + strconv.Itoa(env.gensymCounter) + "__auto__"
			symbolNameMap[item.Value] = newName
			env.gensymCounter++
		}
		return value.NewSymbol(newName), nil
	default:
		return item, nil
	}
}

// essential syntax: let <bindings> <body>
//
// Two forms are supported. The first is the standard let form:
//==============================================================================
// Syntax: <bindings> should have the form:
//
// ((<symbol 1> <init 1>) ...),
//
// where each <init> is an expression, and <body> should be a sequence
// of one or more expressions. It is an error for a <symbol> to
// appear more than once in the list of symbols being bound.
//
// Semantics: The <init>s are evaluated in the current environment (in
// some unspecified order), the <symbol>s are bound to fresh
// locations holding the results, the <body> is evaluated in the
// extended environment, and the value of the last expression of
// <body> is returned. Each binding of a <symbol> has <body> as its
// region.
//==============================================================================
// The second form is the clojure
// (https://clojuredocs.org/clojure.core/let) let form:
//
// Syntax: <bindings> should have the form:
//
// [<symbol 1> <init 1> ... <symbol n> <init n>]
//
// Semantics: Semantics are as above, except that the <init>s are
// evaluated in order, and all preceding <init>s are available to
// subsequent <init>s.
func (env *environment) evalLet(n *value.List) (value.Value, value.Continuation, error) {
	items := listAsSlice(n)
	if len(items) < 3 {
		return nil, nil, env.errorf(n, "invalid let, need bindings and body")
	}

	var bindingsMap map[string]value.Value
	var err error
	switch bindings := items[1].(type) {
	case *value.List:
		bindingsMap, err = env.evalListBindings(bindings)
	case *value.Vector:
		bindingsMap, err = env.evalVectorBindings(bindings)
	default:
		return nil, nil, env.errorf(items[1], "invalid let, bindings must be a list or vector")
	}
	if err != nil {
		return nil, nil, err
	}

	// create a new environment with the bindings
	newEnv := env.PushScope().(*environment)
	for name, val := range bindingsMap {
		newEnv.Define(name, val)
	}

	// evaluate the body
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

func (env *environment) evalListBindings(bindingList *value.List) (map[string]value.Value, error) {
	bindings := listAsSlice(bindingList)

	// shuffle the bindings to evaluate them in a random order. this
	// prevents users from relying on the order of evaluation.
	shuffled := make([]*value.List, len(bindings))
	for i, j := range rand.Perm(len(bindings)) {
		item, ok := bindings[i].(*value.List)
		if !ok || item.Count() != 2 {
			return nil, env.errorf(bindings[i], "invalid let, bindings must be a list of lists of length 2")
		}
		shuffled[j] = item
	}

	// evaluate the bindings in a random order
	bindingsMap := make(map[string]value.Value)
	for _, binding := range shuffled {
		nameValue := binding.Item()
		name, ok := nameValue.(*value.Symbol)
		if !ok {
			return nil, env.errorf(nameValue, "invalid let, binding name must be a symbol")
		}
		if _, ok := bindingsMap[name.Value]; ok {
			return nil, env.errorf(nameValue, "invalid let, duplicate binding name")
		}
		val, err := env.Eval(binding.Next().Item())
		if err != nil {
			return nil, err
		}
		bindingsMap[name.Value] = val
	}
	return bindingsMap, nil
}

func (env *environment) evalVectorBindings(bindings *value.Vector) (map[string]value.Value, error) {
	if bindings.Count()%2 != 0 {
		return nil, env.errorf(bindings, "invalid let, bindings must be a vector of even length")
	}

	newEnv := env.PushScope().(*environment)
	bindingsMap := make(map[string]value.Value)
	for i := 0; i < bindings.Count(); i += 2 {
		nameValue := bindings.ValueAt(i)
		name, ok := nameValue.(*value.Symbol)
		if !ok {
			return nil, env.errorf(nameValue, "invalid let, binding name must be a symbol")
		}
		if _, ok := bindingsMap[name.Value]; ok {
			return nil, env.errorf(nameValue, "invalid let, duplicate binding name")
		}
		val, err := newEnv.Eval(bindings.ValueAt(i + 1))
		if err != nil {
			return nil, err
		}
		bindingsMap[name.Value] = val
		newEnv.Define(name.Value, val)
	}

	return bindingsMap, nil
}

func (env *environment) evalDefMacro(n *value.List) (value.Value, error) {
	fn, err := env.evalFn(n)
	if err != nil {
		return nil, err
	}

	sym, ok := n.Next().Item().(*value.Symbol)
	if !ok {
		return nil, env.errorf(n.Next().Item(), "invalid defmacro, name must be a symbol")
	}

	env.DefineMacro(sym.Value, fn.(*value.Func))
	return nil, nil
}

func (env *environment) applyMacro(fn *value.Func, argList *value.List) (value.Value, value.Continuation, error) {
	args := listAsSlice(argList)
	res, c, err := env.applyFunc(fn, args)
	if err != nil {
		return nil, nil, err
	}
	if c == nil {
		return nil, func() (value.Value, value.Continuation, error) {
			return env.eval(res)
		}, nil
	}

	// continue evaluating until we get a result
	for c != nil && err == nil {
		res, c, err = c()
	}
	if err != nil {
		return nil, nil, err
	}

	return nil, func() (value.Value, value.Continuation, error) {
		return env.eval(res)
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
