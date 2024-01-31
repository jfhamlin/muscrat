package stochastic

import (
	"context"
	"math/rand"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewRRand(rnd *rand.Rand) ugen.UGen {
	var val float64
	inited := false
	lastTrig := 0.0
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		trigs := cfg.InputSamples["trig"]
		mins := cfg.InputSamples["min"]
		maxs := cfg.InputSamples["max"]

		if !inited {
			val = mins[0] + rnd.Float64()*(maxs[0]-mins[0])
			inited = true
		}
		for i := range out {
			if trigs[i] > 0 && lastTrig <= 0 {
				val = mins[i] + rnd.Float64()*(maxs[i]-mins[i])
			}
			out[i] = val
		}
	})
}
