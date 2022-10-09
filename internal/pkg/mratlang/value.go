package mratlang

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nsf/sexp"

	"github.com/jfhamlin/muscrat/internal/pkg/graph"
)

type Value interface {
	String() string
	Equal(Value) bool
}

// List is a list of values.
type List struct {
	Values []Value
}

func (l *List) String() string {
	builder := strings.Builder{}
	builder.WriteString("(")
	for i, v := range l.Values {
		builder.WriteString(v.String())
		if i < len(l.Values)-1 {
			builder.WriteString(" ")
		}
	}
	builder.WriteString(")")
	return builder.String()
}

func (l *List) Equal(v Value) bool {
	other, ok := v.(*List)
	if !ok {
		return false
	}
	if len(l.Values) != len(other.Values) {
		return false
	}
	for i, v := range l.Values {
		if !v.Equal(other.Values[i]) {
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
	return fmt.Sprintf("Gen{%v}", g.NodeID)
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
	Value bool
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
	Value float64
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
	Value string
}

func (s *Str) String() string {
	return s.Value
}

func (s *Str) Equal(v Value) bool {
	other, ok := v.(*Str)
	if !ok {
		return false
	}
	return s.Value == other.Value
}

// Func is a function.
type Func struct {
	lambdaName string
	variadic   bool
	argNames   []string
	env        *environment
	node       *sexp.Node
}

func (f *Func) String() string {
	// TODO: it would be nice to print something about where the function
	// was defined.
	return fmt.Sprintf("Func{%v}", f.node)
}

func (f *Func) Equal(v Value) bool {
	other, ok := v.(*Func)
	if !ok {
		return false
	}
	return f.node == other.node
}

func (f *Func) Apply(env *environment, args []Value) (Value, error) {
	fnEnv := f.env.pushScope()
	fnEnv.define("$args", &List{Values: args})
	if f.lambdaName != "" {
		// Define the function name in the environment.
		fnEnv.define(f.lambdaName, f)
	}

	for i, argName := range f.argNames {
		if i >= len(args) {
			return nil, fmt.Errorf("too few arguments to function")
		}
		fnEnv.define(argName, args[i])
	}
	if f.variadic {
		for i := len(f.argNames); i < len(args); i++ {
			fnEnv.define(fmt.Sprintf("$%d", i), args[i])
		}
	}

	return fnEnv.evalNode(f.node)
}

// BuiltinFunc is a builtin function.
type BuiltinFunc struct {
	applyer
	name     string
	variadic bool
	argNames []string
}

func (f *BuiltinFunc) String() string {
	return fmt.Sprintf("*builtin %s*", f.name)
}

func (f *BuiltinFunc) Equal(v Value) bool {
	other, ok := v.(*BuiltinFunc)
	if !ok {
		return false
	}
	return f == other
}
