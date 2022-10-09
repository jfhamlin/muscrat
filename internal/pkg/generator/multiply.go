package generator

import "context"

func NewProduct() SampleGenerator {
	return SampleGeneratorFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := range res {
			res[i] = 1
			for _, s := range cfg.InputSamples {
				res[i] *= s[i]
			}
		}
		return res
	})
}
