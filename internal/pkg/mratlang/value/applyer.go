package value

type Applyer interface {
	Apply(env Environment, args []Value) (Value, error)
}

type ApplyerFunc func(env Environment, args []Value) (Value, error)

func (f ApplyerFunc) Apply(env Environment, args []Value) (Value, error) {
	return f(env, args)
}
