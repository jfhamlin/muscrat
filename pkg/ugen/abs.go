package ugen

import (
	"context"
	"math"
)

func NewAbs() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		res := make([]float64, n)
		in := cfg.InputSamples["in"]
		for i := 0; i < n; i++ {
			res[i] = math.Abs(in[i])
		}
		return res
	})
}
