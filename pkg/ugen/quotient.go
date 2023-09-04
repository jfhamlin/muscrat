package ugen

import "context"

func NewQuotient() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		xs := CollectIndexedInputs(cfg)
		if len(xs) == 0 {
			return
		}
		for i := range out {
			out[i] = xs[0][i]
			for _, x := range xs[1:] {
				out[i] /= x[i]
			}
		}
	})
}
