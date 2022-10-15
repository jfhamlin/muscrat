package value

import (
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
	Items []Value
}

func NewList(values []Value, opts ...Option) *List {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return &List{
		Section: o.section,
		Items:   values,
	}
}

func (l *List) String() string {
	b := strings.Builder{}

	// special case for quoted values
	if len(l.Items) == 2 {
		// TODO: only do this if it used quote shorthand when read.
		if sym, ok := l.Items[0].(*Symbol); ok && sym.Value == "quote" {
			b.WriteString("'")
			b.WriteString(l.Items[1].String())
			return b.String()
		}
	}

	b.WriteString("(")
	for i, v := range l.Items {
		b.WriteString(v.String())
		if i < len(l.Items)-1 {
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
	if len(l.Items) != len(other.Items) {
		return false
	}
	for i, v := range l.Items {
		if !v.Equal(other.Items[i]) {
			return false
		}
	}
	return true
}

// Gen is a generator.
type Gen struct {
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

func (g *Gen) Pos() ast.Pos {
	return ast.Pos{}
}

// Bool is a boolean value.
type Bool struct {
	Value bool
}

func NewBool(b bool) *Bool {
	return &Bool{b}
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

func (b *Bool) Pos() ast.Pos {
	return ast.Pos{}
}

// Num is a number.
type Num struct {
	Value float64
}

func NewNum(v float64) *Num {
	return &Num{Value: v}
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

func (n *Num) Pos() ast.Pos {
	return ast.Pos{}
}

// Str is a string.
type Str struct {
	Value string
}

func NewStr(s string) *Str {
	return &Str{Value: s}
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

func (s *Str) Pos() ast.Pos {
	return ast.Pos{}
}

// Keyword represents a keyword. Syyntactically, a keyword is a symbol
// that starts with a colon and evaluates to itself.
type Keyword struct {
	Value string
}

func NewKeyword(s string) *Keyword {
	return &Keyword{Value: s}
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

func (k *Keyword) Pos() ast.Pos {
	return ast.Pos{}
}

// Func is a function.
type Func struct {
	LambdaName string
	Variadic   bool
	ArgNames   []string
	Env        Environment
	Exprs      *List
}

func (f *Func) String() string {
	return fmt.Sprintf("(fn (%v) %s)", f.ArgNames, f.Exprs)
}

func (f *Func) Equal(v Value) bool {
	other, ok := v.(*Func)
	if !ok {
		return false
	}
	return f.Exprs == other.Exprs
}

func (f *Func) Pos() ast.Pos {
	return ast.Pos{}
}

func (f *Func) Apply(env Environment, args []Value) (Value, error) {
	fnEnv := f.Env.PushScope()
	fnEnv.Define("$args", &List{Items: args})
	if f.LambdaName != "" {
		// Define the function name in the environment.
		fnEnv.Define(f.LambdaName, f)
	}

	for i, argName := range f.ArgNames {
		if i >= len(args) {
			return nil, fmt.Errorf("too few arguments to function")
		}
		fnEnv.Define(argName, args[i])
	}
	if f.Variadic {
		for i := len(f.ArgNames); i < len(args); i++ {
			fnEnv.Define(fmt.Sprintf("$%d", i), args[i])
		}
	}

	var res Value
	for _, expr := range f.Exprs.Items {
		v, err := fnEnv.Eval(expr)
		if err != nil {
			return nil, err
		}
		res = v
	}
	return res, nil
}

// BuiltinFunc is a builtin function.
type BuiltinFunc struct {
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

func (f *BuiltinFunc) Pos() ast.Pos {
	return ast.Pos{}
}

type Symbol struct {
	Value string
}

func NewSymbol(s string) *Symbol {
	return &Symbol{Value: s}
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

func (s *Symbol) Pos() ast.Pos {
	return ast.Pos{}
}
