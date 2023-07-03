package ugen

import (
	"context"
	"math"
)

func NewExp() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		res := make([]float64, n)
		in := cfg.InputSamples["in"]
		for i := 0; i < n; i++ {
			res[i] = math.Exp(in[i])
		}
		return res
	})
}
