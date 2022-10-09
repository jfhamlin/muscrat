package mratlang

type scope struct {
	parent *scope
	syms   map[string]Value
}

func newScope() *scope {
	return &scope{syms: make(map[string]Value)}
}

func (s *scope) define(name string, val Value) {
	s.syms[name] = val
}

func (s *scope) push() *scope {
	return &scope{parent: s, syms: make(map[string]Value)}
}

func (s *scope) lookup(name string) (Value, bool) {
	if v, ok := s.syms[name]; ok {
		return v, true
	}
	if s.parent == nil {
		return nil, false
	}
	return s.parent.lookup(name)
}
