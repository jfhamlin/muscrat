package ugen

import (
	"context"
	"math"
)

func NewCopySign() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		sign := cfg.InputSamples["sign"]
		for i := range out {
			out[i] = math.Copysign(in[i], sign[i])
		}
	})
}
