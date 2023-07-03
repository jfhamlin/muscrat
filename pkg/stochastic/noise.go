package stochastic

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

// NewNoise returns a new Noise ugen. If freq is 0, the noise will be
// white.
func NewNoise(opts ...ugen.Option) ugen.SampleGenerator {
	o := ugen.DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	rnd := o.Rand
	add := o.Add
	mul := o.Mul

	last := mul*2*rnd.Float64() - 1 + add
	counter := 0
	// Logic taken from supercollider LFNoise0 ugen
	return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		ws := cfg.InputSamples["w"]
		res := make([]float64, n)
		remain := n
		i := 0
		for {
			freq := 0.0
			if len(ws) > 0 {
				freq = ws[i]
			}
			freq = math.Max(0, freq)
			if freq == 0 {
				counter = 1
				last = mul*(2*rnd.Float64()-1) + add
			} else if counter <= 0 {
				counter = int(float64(cfg.SampleRateHz) / freq)
				if counter <= 0 {
					counter = 1
				}
				last = mul*(2*rnd.Float64()-1) + add
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
