package generator

import "context"

func NewConstant(val float64) SampleGenerator {
	var buf []float64
	return SampleGeneratorFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		if len(buf) < n {
			buf = make([]float64, n)
			for i := 0; i < n; i++ {
				buf[i] = val
			}
		}
		return buf[:n]
	})
}
