package ugen

import "context"

func NewSum() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		for _, s := range cfg.InputSamples {
			// index the last element of s to lift the bounds check
			_ = s[len(out)-1]
			for i := range out {
				out[i] += s[i]
			}
		}
	})
}
