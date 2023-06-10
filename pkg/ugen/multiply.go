package ugen

import "context"

func NewProduct() SampleGenerator {
	return SampleGeneratorFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := 0; i < n; i++ {
			res[i] = 1
		}
		for _, s := range cfg.InputSamples {
			for i := range res {
				res[i] *= s[i]
			}
		}
		return res
	})
}
