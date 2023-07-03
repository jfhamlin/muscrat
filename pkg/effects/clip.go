package effects

import (
	"context"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewClip() ugen.UGen {
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		in := cfg.InputSamples["in"]
		los := cfg.InputSamples["lo"]
		his := cfg.InputSamples["hi"]

		res := make([]float64, n)
		for i := 0; i < n; i++ {
			x := in[i]
			lo := los[i]
			hi := his[i]
			if x < lo {
				x = lo
			}
			if x > hi {
				x = hi
			}
			res[i] = x
		}
		return res
	})
}
