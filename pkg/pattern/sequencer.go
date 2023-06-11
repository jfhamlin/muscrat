package pattern

import (
	"context"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewSequencer() ugen.UGen {
	index := 0
	lastTrig := 0.0
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		trigs := cfg.InputSamples["trigger"]
		freqs := ugen.CollectIndexedInputs(cfg)
		out := make([]float64, n)
		for i := 0; i < n; i++ {
			if trigs[i] > 0.0 && lastTrig <= 0.0 {
				index = (index + 1) % len(freqs)
			}
			out[i] = freqs[index][i]
			lastTrig = trigs[i]
		}
		return out
	})
}
