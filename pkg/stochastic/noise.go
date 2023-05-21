package stochastic

import (
	"context"
	"math"
	"math/rand"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

// NewNoise returns a new Noise ugen. If freq is 0, the noise will be
// white.
func NewNoise(freq float64, opts ...Option) ugen.SampleGenerator {
	o := options{
		rand: rand.New(rand.NewSource(0)),
		add:  0.0,
		mul:  1.0,
	}
	for _, opt := range opts {
		opt(&o)
	}

	rnd := o.rand
	add := o.add
	mul := o.mul

	freq = math.Max(0, freq)

	if freq == 0 {
		return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
			res := make([]float64, n)
			for i := 0; i < n; i++ {
				res[i] = mul*2*rnd.Float64() - 1 + add
			}
			return res
		})
	}

	last := mul*2*rnd.Float64() - 1 + add
	counter := 0
	// Logic taken from supercollider LFNoise0 ugen
	return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		remain := n
		i := 0
		for {
			if counter <= 0 {
				counter = int(float64(cfg.SampleRateHz) / freq)
				if counter <= 0 {
					counter = 1
				}
				last = mul*2*rnd.Float64() - 1 + add
			}
			nsamp := counter
			if nsamp > remain {
				nsamp = remain
			}
			remain -= nsamp
			counter -= nsamp
			for j := 0; j < nsamp; j++ {
				res[i] = last
				i++
			}
			if remain <= 0 {
				break
			}
		}
		return res
	})
}