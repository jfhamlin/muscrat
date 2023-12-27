package ugen

import (
	"context"
	"math"
)

func NewSine() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		for _, s := range cfg.InputSamples {
			for i := range out {
				out[i] += math.Sin(s[i])
			}
		}
	})
}
