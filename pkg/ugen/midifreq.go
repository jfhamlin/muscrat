package ugen

import (
	"context"
	"math"
)

func NewMIDIFreq() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		res := make([]float64, n)
		in := cfg.InputSamples["in"]
		for i := 0; i < n; i++ {
			res[i] = 440 * math.Pow(2, (in[i]-69)/12)
		}
		return res
	})
}
