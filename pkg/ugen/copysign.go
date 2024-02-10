package ugen

import (
	"context"
	"math"
)

func NewCopySign() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		sign := cfg.InputSamples["sign"]
		// index the last element of in and sign to lift the bounds check
		_ = in[len(out)-1]
		_ = sign[len(out)-1]
		for i := range out {
			out[i] = math.Copysign(in[i], sign[i])
		}
	})
}
