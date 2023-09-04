package ugen

import (
	"context"
	"math"
)

func NewAbs() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		for i := range out {
			out[i] = math.Abs(in[i])
		}
	})
}
