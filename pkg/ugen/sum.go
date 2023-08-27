package ugen

import "context"

func NewSum() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		for _, s := range cfg.InputSamples {
			for i := range out {
				out[i] += s[i]
			}
		}
	})
}
