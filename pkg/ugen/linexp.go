package ugen

import (
	"context"
	"math"
)

func NewLinExp() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		srclo := cfg.InputSamples["srclo"]
		srchi := cfg.InputSamples["srchi"]
		dstlo := cfg.InputSamples["dstlo"]
		dsthi := cfg.InputSamples["dsthi"]

		_ = in[len(out)-1]
		_ = srclo[len(out)-1]
		_ = srchi[len(out)-1]
		_ = dstlo[len(out)-1]
		_ = dsthi[len(out)-1]

		for i := range out {
			rsrcRange := 1 / (srchi[i] - srclo[i])
			rminusLo := rsrcRange * -srclo[i]
			dhi := dsthi[i]
			dlo := dstlo[i]
			dstRatio := dhi / dlo
			out[i] = dlo * math.Pow(dstRatio, in[i]*rsrcRange+rminusLo)
		}
	})
}
