package ugen

import (
	"context"
	"math"
)

func NewCopySign() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		res := make([]float64, n)
		in := cfg.InputSamples["in"]
		sign := cfg.InputSamples["sign"]
		for i := 0; i < n; i++ {
			res[i] = math.Copysign(in[i], sign[i])
		}
		return res
	})
}
