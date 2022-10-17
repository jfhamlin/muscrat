package wavtabs

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
)

func Generator(wavtab Table) generator.SampleGenerator {
	phase := 0.0
	lastSync := 0.0
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		ws := cfg.InputSamples["w"]
		phases := cfg.InputSamples["phase"]
		syncs := cfg.InputSamples["sync"]

		res := make([]float64, n)
		w := 440.0 // default frequency
		for i := 0; i < n; i++ {
			if i < len(phases) {
				phase = phases[i]
			}
			res[i] = wavtab.Lerp(phase)

			if i < len(ws) {
				w = ws[i]
			}
			phase += w / float64(cfg.SampleRateHz)
			// keep phase in [0, 1)
			phase -= math.Floor(phase)

			// sync on the falling edge of the sync input if present
			if i < len(syncs) && syncs[i] < lastSync {
				phase = 0.0
				lastSync = syncs[i]
			}
		}
		return res
	})
}
