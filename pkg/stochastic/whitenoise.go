package stochastic

import (
	"context"
	"math/rand"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewNoise(opts ...Option) ugen.SampleGenerator {
	o := options{
		rand: rand.New(rand.NewSource(0)),
	}
	for _, opt := range opts {
		opt(&o)
	}

	rnd := o.rand

	return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := 0; i < n; i++ {
			res[i] = 2*rnd.Float64() - 1
		}
		return res
	})
}
