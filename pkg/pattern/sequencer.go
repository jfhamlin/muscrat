package pattern

import (
	"context"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewSequencer() ugen.UGen {
	index := 0
	lastTrig := 1.0
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		trigs := cfg.InputSamples["trigger"]
		vals := ugen.CollectIndexedInputs(cfg)
		if len(vals) == 0 {
			return
		}
		if index >= len(vals) {
			index = 0
		}
		for i := range out {
			if trigs[i] > 0.0 && lastTrig <= 0.0 {
				index = (index + 1) % len(vals)
			}
			out[i] = vals[index][i]
			lastTrig = trigs[i]
		}
	})
}
