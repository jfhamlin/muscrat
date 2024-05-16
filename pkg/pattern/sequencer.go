package pattern

import (
	"context"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewSequencer() ugen.UGen {
	index := 0
	lastTrig := 1.0
	lastSync := 1.0
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		trigs := cfg.InputSamples["trigger"]
		syncs := cfg.InputSamples["sync"]
		if len(syncs) == 0 {
			syncs = ugen.Zeros
		}
		_ = syncs[len(out)-1]

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
			if syncs[i] > 0 && lastSync <= 0 {
				index = 0 // this case before or after the increment case?
			}
			out[i] = vals[index][i]
			lastTrig = trigs[i]
			lastSync = syncs[i]
		}
	})
}
