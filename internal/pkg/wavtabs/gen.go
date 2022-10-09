package wavtabs

import (
	"context"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
)

func Generator(wavtab Table) generator.SampleGenerator {
	phase := 0.0
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		ws := cfg.InputSamples["w"]
		res := make([]float64, n)
		// default frequency; use the last value if we run out of
		// input samples
		w := 0.0
		for i := 0; i < n; i++ {
			if i < len(ws) {
				w = ws[i]
			}
			res[i] = wavtab.Lerp(phase)
			phase += w / float64(cfg.SampleRateHz)
			if phase > 1 {
				phase -= 1
			}
		}
		return res
	})
}
