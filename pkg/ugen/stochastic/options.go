package stochastic

import "math/rand"

type options struct {
	rand *rand.Rand
}

// Option is a function that configures the stochastic generator.
type Option func(*options)

// WithRand sets the random number generator to use.
func WithRand(r *rand.Rand) Option {
	return func(o *options) {
		o.rand = r
	}
}
