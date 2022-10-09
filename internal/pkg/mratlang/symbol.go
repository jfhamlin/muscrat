package mratlang

type Location struct {
}

type Symbol struct {
	Name string
	Help string
	// where the symbol is defined
	// if nil, it is a builtin
	DefLocation *Location
	Value       Value
}

type Package struct {
	Name    string
	Symbols []*Symbol
}
