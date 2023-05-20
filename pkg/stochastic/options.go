package stochastic

import "math/rand"

type options struct {
	rand *rand.Rand
	add  float64
	mul  float64
}

// Option is a function that configures the stochastic generator.
type Option func(*options)

// WithRand sets the random number generator to use.
func WithRand(r *rand.Rand) Option {
	return func(o *options) {
		o.rand = r
	}
}

func WithAdd(add float64) Option {
	return func(o *options) {
		o.add = add
	}
}

func WithMul(mul float64) Option {
	return func(o *options) {
		o.mul = mul
	}
}
