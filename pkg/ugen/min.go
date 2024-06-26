package ugen

import (
	"context"
	"math"
)

func NewMin() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		if len(cfg.InputSamples) == 0 {
			return
		}

		for i := range out {
			out[i] = math.Inf(1)
		}
		for _, s := range cfg.InputSamples {
			// index the last element of s to lift the bounds check
			_ = s[len(out)-1]
			for i := range out {
				if s[i] < out[i] {
					out[i] = s[i]
				}
			}
		}
	})
}
