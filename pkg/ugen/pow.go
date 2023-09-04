package ugen

import (
	"context"
	"math"
)

func NewPow() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		base := cfg.InputSamples["base"]
		exp := cfg.InputSamples["exp"]
		for i := range out {
			out[i] = math.Pow(base[i], exp[i])
		}
	})
}
