package ugen

import (
	"context"
	"math"
)

func NewTanh() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		// index the last element of in to lift the bounds check
		_ = in[len(out)-1]
		for i := range out {
			out[i] = math.Tanh(in[i])
		}
	})
}
