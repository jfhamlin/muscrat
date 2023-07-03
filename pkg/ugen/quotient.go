package ugen

import "context"

func NewQuotient() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		res := make([]float64, n)

		xs := CollectIndexedInputs(cfg)
		if len(xs) == 0 {
			return res
		}
		for i := 0; i < n; i++ {
			res[i] = xs[0][i]
			for _, x := range xs[1:] {
				res[i] /= x[i]
			}
		}
		return res
	})
}
