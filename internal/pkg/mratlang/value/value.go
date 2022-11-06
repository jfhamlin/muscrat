package value

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/ast"
)

// Value is the interface that all values in the language implement.
type Value interface {
	String() string
	Equal(Value) bool

	// Pos returns the position in the source code where the value was
	// created or defined.
	Pos() ast.Pos
}

// Enumerable is an interface for compound values that support
// enumeration.
type Enumerable interface {
	// Enumerate returns a channel that will yield all of the values
	// in the compound value.
	Enumerate() (values <-chan Value, cancel func())
}

// EnumerableFunc is a function that implements the Enumerable
// interface.
type EnumerableFunc func() (<-chan Value, func())

func (f EnumerableFunc) Enumerate() (<-chan Value, func()) {
	return f()
}

// EnumerateAll returns all values in the sequence. If the sequence is
// infinite, this will never return unless the context is cancelled.
func EnumerateAll(ctx context.Context, e Enumerable) ([]Value, error) {
	ch, cancel := e.Enumerate()
	defer cancel()

	var values []Value
	for {
		select {
		case v, ok := <-ch:
			if !ok {
				return values, nil
			}
			values = append(values, v)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Counter is an interface for compound values whose elements can be
// counted.
type Counter interface {
	Count() int
}

type options struct {
	// where the value was defined
	section ast.Section
}

// Option represents an option that can be passed to Value
// constructors.
type Option func(*options)

// WithSection returns an Option that sets the section of the value.
func WithSection(s ast.Section) Option {
	return func(o *options) {
		o.section = s
	}
}

// List is a list of values.
type List struct {
	ast.Section

	// the empty list is represented by a nil item and a nil next. all
	// other lists have a non-nil item and a non-nil next.
	item Value
	next *List
}

var emptyList = &List{}

func NewList(values []Value, opts ...Option) *List {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	list := emptyList
	for i := len(values) - 1; i >= 0; i-- {
		list = &List{
			Section: o.section,
			item:    values[i],
			next:    list,
		}
	}
	return list
}

func ConsList(item Value, next *List) *List {
	if next == nil {
		next = emptyList
	}
	return &List{
		item: item,
		next: next,
	}
}

// Item returns the data from this list node. AKA car.
func (l *List) Item() Value {
	if l.IsEmpty() {
		panic("cannot get item of empty list")
	}
	return l.item
}

// Next returns the next list node. AKA cdr, with the requirement that
// it must be a list.
func (l *List) Next() *List {
	if l.IsEmpty() {
		panic("cannot get next of empty list")
	}
	return l.next
}

func (l *List) IsEmpty() bool {
	return l.item == nil && l.next == nil
}

func (l *List) Count() int {
	count := 0
	for !l.IsEmpty() {
		count++
		l = l.next
	}
	return count
}

func (l *List) Enumerate() (<-chan Value, func()) {
	return enumerateFunc(func() (v Value, ok bool) {
		if l.IsEmpty() {
			return nil, false
		}
		v = l.item
		l = l.next
		return v, true
	})
}

func enumerateItems(items []Value) (<-chan Value, func()) {
	ch := make(chan Value)

	done := make(chan struct{})
	cancel := func() {
		close(done)
	}
	go func() {
		for _, v := range items {
			select {
			case ch <- v:
			case <-done:
				return
			}
		}
		close(ch)
	}()
	return ch, cancel
}

func enumerateFunc(next func() (v Value, ok bool)) (<-chan Value, func()) {
	ch := make(chan Value)

	done := make(chan struct{})
	cancel := func() {
		close(done)
	}
	go func() {
		for {
			v, ok := next()
			if !ok {
				break
			}
			select {
			case ch <- v:
			case <-done:
				return
			}
		}
		close(ch)
	}()
	return ch, cancel
}

func (l *List) String() string {
	b := strings.Builder{}

	// special case for quoted values
	if l.Count() == 2 {
		// TODO: only do this if it used quote shorthand when read.
		if sym, ok := l.item.(*Symbol); ok && sym.Value == "quote" {
			b.WriteString("'")
			b.WriteString(l.next.item.String())
			return b.String()
		}
	}

	b.WriteString("(")
	for cur := l; !cur.IsEmpty(); cur = cur.next {
		v := cur.item
		if v == nil {
			b.WriteString("()")
		} else {
			b.WriteString(v.String())
		}
		if !cur.next.IsEmpty() {
			b.WriteString(" ")
		}
	}
	b.WriteString(")")
	return b.String()
}

func (l *List) Equal(v Value) bool {
	other, ok := v.(*List)
	if !ok {
		return false
	}

	for {
		if l.IsEmpty() != other.IsEmpty() {
			return false
		}
		if l.IsEmpty() {
			return true
		}
		if !l.item.Equal(other.item) {
			return false
		}
		l = l.next
		other = other.next
	}

	return true
}

// Vector is a vector of values.
type Vector struct {
	ast.Section
	Items []Value
}

func NewVector(values []Value, opts ...Option) *Vector {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return &Vector{
		Section: o.section,
		Items:   values,
	}
}

func (v *Vector) Count() int {
	return len(v.Items)
}

func (v *Vector) Enumerate() (<-chan Value, func()) {
	return enumerateItems(v.Items)
}

func (v *Vector) String() string {
	b := strings.Builder{}

	b.WriteString("[")
	for i, el := range v.Items {
		if el == nil {
			b.WriteString("()")
		} else {
			b.WriteString(el.String())
		}
		if i < len(v.Items)-1 {
			b.WriteString(" ")
		}
	}
	b.WriteString("]")
	return b.String()
}

func (v *Vector) Equal(v2 Value) bool {
	other, ok := v2.(*Vector)
	if !ok {
		return false
	}
	if len(v.Items) != len(other.Items) {
		return false
	}
	for i, v := range v.Items {
		if !v.Equal(other.Items[i]) {
			return false
		}
	}
	return true
}

func (v *Vector) Apply(env Environment, args []Value) (Value, error) {
	if len(args) > 2 {
		return nil, fmt.Errorf("vector apply takes one or two arguments")
	}

	index, ok := args[0].(*Num)
	if !ok {
		return nil, fmt.Errorf("vector apply takes a number as an argument")
	}

	i := int(index.Value)
	if i < 0 || i >= len(v.Items) && len(args) == 1 {
		return nil, fmt.Errorf("index out of bounds")
	}
	if i >= len(v.Items) {
		return args[1], nil
	}
	return v.Items[i], nil
}

// Seq is a lazy sequence of values.
type Seq struct {
	Enumerable
}

func (s *Seq) Equal(v Value) bool {
	other, ok := v.(*Seq)
	if !ok {
		return false
	}
	e1, cancel1 := s.Enumerate()
	defer cancel1()
	e2, cancel2 := other.Enumerate()
	defer cancel2()
	for {
		v1, ok1 := <-e1
		v2, ok2 := <-e2
		if ok1 != ok2 {
			return false
		}
		if !ok1 {
			return true
		}
		if !v1.Equal(v2) {
			return false
		}
	}
	return true
}

func (s *Seq) Pos() ast.Pos {
	return ast.Pos{}
}

func (s *Seq) String() string {
	b := strings.Builder{}
	b.WriteString("(")
	e, cancel := s.Enumerate()
	defer cancel()
	first := true
	for {
		v, ok := <-e
		if !ok {
			break
		}
		if !first {
			b.WriteString(" ")
		}
		first = false
		b.WriteString(v.String())
	}
	b.WriteString(")")
	return b.String()
}

// Gen is a generator.
type Gen struct {
	ast.Section
	NodeID graph.NodeID
}

func (g *Gen) String() string {
	return fmt.Sprintf("(<gen>)")
}

func (g *Gen) Equal(v Value) bool {
	other, ok := v.(*Gen)
	if !ok {
		return false
	}
	return g.NodeID == other.NodeID
}

// Bool is a boolean value.
type Bool struct {
	ast.Section
	Value bool
}

func NewBool(b bool, opts ...Option) *Bool {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return &Bool{
		Section: o.section,
		Value:   b,
	}
}

func (b *Bool) String() string {
	if b.Value {
		return "#t"
	}
	return "#f"
}

func (b *Bool) Equal(v Value) bool {
	other, ok := v.(*Bool)
	if !ok {
		return false
	}
	return b.Value == other.Value
}

// Num is a number.
type Num struct {
	ast.Section
	Value float64
}

func NewNum(n float64, opts ...Option) *Num {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return &Num{
		Section: o.section,
		Value:   n,
	}
}

func (n *Num) String() string {
	return strconv.FormatFloat(n.Value, 'f', -1, 64)
}

func (n *Num) Equal(v Value) bool {
	other, ok := v.(*Num)
	if !ok {
		return false
	}
	return n.Value == other.Value
}

// Str is a string.
type Str struct {
	ast.Section
	Value string
}

func NewStr(s string, opts ...Option) *Str {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return &Str{
		Section: o.section,
		Value:   s,
	}
}

func (s *Str) String() string {
	return "\"" + s.Value + "\""
}

func (s *Str) Equal(v Value) bool {
	other, ok := v.(*Str)
	if !ok {
		return false
	}
	return s.Value == other.Value
}

// Keyword represents a keyword. Syyntactically, a keyword is a symbol
// that starts with a colon and evaluates to itself.
type Keyword struct {
	ast.Section
	Value string
}

func NewKeyword(s string, opts ...Option) *Keyword {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return &Keyword{
		Section: o.section,
		Value:   s,
	}
}

func (k *Keyword) String() string {
	return ":" + k.Value
}

func (k *Keyword) Equal(v Value) bool {
	other, ok := v.(*Keyword)
	if !ok {
		return false
	}
	return k.Value == other.Value
}

// Func is a function.
type Func struct {
	ast.Section
	LambdaName string
	Variadic   bool
	ArgNames   []string
	Env        Environment
	Exprs      *List
}

func (f *Func) String() string {
	b := strings.Builder{}
	b.WriteString("(fn")
	if f.LambdaName != "" {
		b.WriteString(" ")
		b.WriteString(f.LambdaName)
	}
	b.WriteString(" (")
	for i, arg := range f.ArgNames {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(arg)
	}
	if f.Variadic {
		if len(f.ArgNames) > 0 {
			b.WriteString(" ")
		}
		b.WriteString("&")
		b.WriteString(f.ArgNames[len(f.ArgNames)-1])
	}
	b.WriteString(") ")
	for cur := f.Exprs; !cur.IsEmpty(); cur = cur.Next() {
		if cur != f.Exprs {
			b.WriteString(" ")
		}
		b.WriteString(cur.Item().String())
	}
	b.WriteString(")")
	return b.String()
}

func (f *Func) Equal(v Value) bool {
	other, ok := v.(*Func)
	if !ok {
		return false
	}
	return f.Exprs == other.Exprs
}

func errorWithStack(err error, stackFrame StackFrame) error {
	if err == nil {
		return nil
	}
	valErr, ok := err.(*Error)
	if !ok {
		return NewError(stackFrame, err)
	}
	return valErr.AddStack(stackFrame)
}

func (f *Func) Apply(env Environment, args []Value) (Value, error) {
	var res Value
	var err error
	continuation := func() (Value, Continuation, error) {
		return f.ContinuationApply(env, args)
	}
	for {
		res, continuation, err = continuation()
		if err != nil {
			return nil, err
		}
		if continuation == nil {
			return res, nil
		}
	}
}

func (f *Func) ContinuationApply(env Environment, args []Value) (Value, Continuation, error) {
	// function name for error messages
	fnName := f.LambdaName
	if fnName == "" {
		fnName = "<anonymous function>"
	}

	fnEnv := f.Env.PushScope()
	fnEnv.Define("$args", &Vector{Items: args})
	if f.LambdaName != "" {
		// Define the function name in the environment.
		fnEnv.Define(f.LambdaName, f)
	}

	for i, argName := range f.ArgNames {
		if i >= len(args) {
			return nil, nil, fmt.Errorf("too few arguments to function")
		}
		fnEnv.Define(argName, args[i])
	}
	if f.Variadic {
		for i := len(f.ArgNames); i < len(args); i++ {
			fnEnv.Define(fmt.Sprintf("$%d", i), args[i])
		}
	}

	var exprs []Value
	for cur := f.Exprs; !cur.IsEmpty(); cur = cur.next {
		exprs = append(exprs, cur.item)
	}
	if len(exprs) == 0 {
		panic("empty function body")
	}

	for _, expr := range exprs[:len(exprs)-1] {
		_, err := fnEnv.Eval(expr)
		if err != nil {
			return nil, nil, errorWithStack(err, StackFrame{
				FunctionName: fnName,
				Pos:          expr.Pos(),
			})
		}
	}
	// return the last expression as a continuation
	lastExpr := exprs[len(exprs)-1]
	return nil, func() (Value, Continuation, error) {
		v, c, err := fnEnv.ContinuationEval(lastExpr)
		if err != nil {
			return nil, nil, errorWithStack(err, StackFrame{
				FunctionName: fnName,
				Pos:          lastExpr.Pos(),
			})
		}
		return v, c, nil
	}, nil
}

// BuiltinFunc is a builtin function.
type BuiltinFunc struct {
	ast.Section
	Applyer
	Name     string
	variadic bool
	argNames []string
}

func (f *BuiltinFunc) String() string {
	return fmt.Sprintf("*builtin %s*", f.Name)
}

func (f *BuiltinFunc) Equal(v Value) bool {
	other, ok := v.(*BuiltinFunc)
	if !ok {
		return false
	}
	return f == other
}

func (f *BuiltinFunc) Apply(env Environment, args []Value) (Value, error) {
	val, err := f.Applyer.Apply(env, args)
	if err != nil {
		return nil, NewError(StackFrame{
			FunctionName: "* builtin " + f.Name + " *",
			Pos:          f.Section.Pos(),
		}, err)
	}
	return val, nil
}

type Symbol struct {
	ast.Section
	Value string
}

func NewSymbol(s string, opts ...Option) *Symbol {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return &Symbol{
		Section: o.section,
		Value:   s,
	}
}

func (s *Symbol) String() string {
	return s.Value
}

func (s *Symbol) Equal(v Value) bool {
	other, ok := v.(*Symbol)
	if !ok {
		return false
	}
	return s.Value == other.Value
}
