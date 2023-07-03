package stochastic

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

// NewNoiseQuad returns a new NoiseQuad ugen, which, as
// Supercollider's Noise2, generates quadratically interpolated random
// values at a rate given by the nearest integer division of the
// sample rate by the freq argument.
func NewNoiseQuad(opts ...ugen.Option) ugen.SampleGenerator {
	o := ugen.DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	rnd := o.Rand
	add := o.Add
	mul := o.Mul

	level := mul*(2*rnd.Float64()-1) + add
	slope := 0.0
	curve := 0.0
	nextValue := 0.0
	nextMidPt := 0.0
	counter := 0

	const defaultFreq = 500.0

	// Logic taken from supercollider LFNoise2 ugen
	return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		ws := cfg.InputSamples["w"]
		res := make([]float64, n)

		remain := n
		i := 0
		for remain > 0 {
			freq := defaultFreq
			if len(ws) > 0 {
				freq = math.Max(ws[i], 0.001)
			}

			if counter <= 0 {
				value := nextValue
				nextValue = mul*(2*rnd.Float64()-1) + add
				level = nextMidPt
				nextMidPt = (value + nextValue) * 0.5

				counter = int(float64(cfg.SampleRateHz) / freq)
				if counter < 2 {
					counter = 2
				}
				fseglen := float64(counter)
				curve = 2.0 * (nextMidPt - level - fseglen*slope) / (fseglen*fseglen + fseglen)
			}
			nsamp := counter
			if nsamp > remain {
				nsamp = remain
			}
			remain -= nsamp
			counter -= nsamp
			for j := 0; j < nsamp; j++ {
				res[i] = level
				slope += curve
				level += slope
				i++
			}
		}
		return res
	})
}
