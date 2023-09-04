package ugen

import (
	"context"
	"math"
)

func NewMIDIFreq() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		for i := range out {
			out[i] = 440 * math.Pow(2, (in[i]-69)/12)
		}
	})
}
