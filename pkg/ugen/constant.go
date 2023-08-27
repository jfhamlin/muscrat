package ugen

import "context"

func NewConstant(val float64) UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		for i := range out {
			out[i] = val
		}
	})
}
