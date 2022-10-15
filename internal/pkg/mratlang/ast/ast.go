package ast

import (
	"strconv"
	"strings"
)

type Node interface {
	Pos() Pos
	End() Pos

	String() string

	// prevent external implementations
	private()
}

type Pos struct {
	Filename string
	Line     int
	Column   int
}

type Section struct {
	StartPos, EndPos Pos
	// TODO: consider adding information about whitespace and comments.
}

func (p Section) Pos() Pos { return p.StartPos }
func (p Section) End() Pos { return p.EndPos }

type List struct {
	Section
	Items []Node
}

func NewList(items []Node, pos Section) *List {
	return &List{Section: pos, Items: items}
}

func (l *List) String() string {
	var b strings.Builder

	// special case for quoted values
	if len(l.Items) == 2 {
		// it's only possible for both the list and the second item to end
		// at the same position if the value used the quote shorthand in
		// the input.
		if sym, ok := l.Items[0].(*Symbol); ok && sym.Value == "quote" && l.End() == l.Items[1].End() {
			b.WriteString("'")
			b.WriteString(l.Items[1].String())
			return b.String()
		}
	}

	b.WriteString("(")
	for i, item := range l.Items {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(item.String())
	}
	b.WriteString(")")
	return b.String()
}

func (l *List) private() {}

type String struct {
	Section
	Value string
}

func NewString(value string, pos Section) *String {
	return &String{Section: pos, Value: value}
}

func (s *String) String() string {
	// TODO: escape characters
	return "\"" + s.Value + "\""
}

func (s *String) private() {}

type Bool struct {
	Section
	Value bool
}

func NewBool(value bool, pos Section) *Bool {
	return &Bool{Section: pos, Value: value}
}

func (b *Bool) String() string {
	if b.Value {
		return "#t"
	}
	return "#f"
}

func (b *Bool) private() {}

type Symbol struct {
	Section
	Value string
}

func NewSymbol(value string, pos Section) *Symbol {
	return &Symbol{Section: pos, Value: value}
}

func (s *Symbol) String() string {
	return s.Value
}

func (s *Symbol) private() {}

type Keyword struct {
	Section
	Value string
}

func NewKeyword(value string, pos Section) *Keyword {
	return &Keyword{Section: pos, Value: value}
}

func (k *Keyword) String() string {
	return ":" + k.Value
}

func (k *Keyword) private() {}

type Number struct {
	Section
	Value float64
}

func NewNumber(value float64, pos Section) *Number {
	return &Number{Section: pos, Value: value}
}

func (n *Number) String() string {
	return strconv.FormatFloat(n.Value, 'f', -1, 64)
}

func (n *Number) private() {}
