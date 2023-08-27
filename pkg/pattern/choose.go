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
		freqs := ugen.CollectIndexedInputs(cfg)
		for i := range out {
			if trigs[i] > 0.0 && lastTrig <= 0.0 {
				index = rand.Intn(len(freqs))
			}
			out[i] = freqs[index][i]
			lastTrig = trigs[i]
		}
	})
}
