package ugen

import "math/rand"

type (
	Interp int

	Options struct {
		Rand             *rand.Rand
		Add              float64
		Mul              float64
		Interp           Interp
		DefaultDutyCycle float64
	}
)

const (
	InterpNone Interp = iota
	InterpLinear
	InterpCubic
)

func DefaultOptions() Options {
	return Options{
		Rand:             rand.New(rand.NewSource(rand.Int63())),
		Add:              0.0,
		Mul:              1.0,
		DefaultDutyCycle: 1.0,
	}
}

// Option is a function that configures a generator.
type Option func(*Options)

// WithSeed sets the seed for the random number generator.
func WithSeed(seed int64) Option {
	return func(o *Options) {
		o.Rand = rand.New(rand.NewSource(seed))
	}
}

// WithRand sets the random number generator to use.
func WithRand(r *rand.Rand) Option {
	return func(o *Options) {
		o.Rand = r
	}
}

func WithAdd(add float64) Option {
	return func(o *Options) {
		o.Add = add
	}
}

func WithMul(mul float64) Option {
	return func(o *Options) {
		o.Mul = mul
	}
}

func WithInterp(interp Interp) Option {
	return func(o *Options) {
		o.Interp = interp
	}
}

func WithDefaultDutyCycle(dc float64) Option {
	return func(o *Options) {
		o.DefaultDutyCycle = dc
	}
}
