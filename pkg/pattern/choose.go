package pattern

import (
	"context"
	"math/rand"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewChoose() ugen.UGen {
	index := 0
	lastTrig := 0.0
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		trigs := cfg.InputSamples["trigger"]
		vals := ugen.CollectIndexedInputs(cfg)
		if len(vals) == 0 {
			return
		}
		// if index is out of bounds, resample
		if index >= len(vals) {
			index = rand.Intn(len(vals))
		}
		for i := range out {
			if trigs[i] > 0.0 && lastTrig <= 0.0 {
				index = rand.Intn(len(vals))
			}
			out[i] = vals[index][i]
			lastTrig = trigs[i]
		}
	})
}
