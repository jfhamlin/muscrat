package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewLowpassFilter() ugen.UGen {
	// Translated from the SuperCollider extension source code here, which in turn mimics the
	// max/msp lores~ object:
	// https://github.com/v7b1/vb_UGens/blob/fea1587dd2165457c4a016214d17216987b56f00/projects/vbUtils/vbUtils.cpp
	var a1, a2, fqterm, resterm, scale, ym1, ym2 float64
	lastCut, lastRes := -1.0, -1.0
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		cuts := cfg.InputSamples["cutoff"]
		ress := cfg.InputSamples["resonance"]

		for i := range out {
			cut := cuts[i]
			res := ress[i]
			// clamp resonance to [0, 1)
			if res < 0 {
				res = 0
			} else if res >= 1 {
				res = 1.0 - 1e-20
			}

			if cut != lastCut || res != lastRes {
				if res != lastRes {
					resterm = math.Exp(res*0.125) * 0.882497
				}
				if cut != lastCut {
					fqterm = math.Cos(cut * math.Pi * 2 / float64(cfg.SampleRateHz))
				}
				// recalculate the coefficients.
				a1 = -2 * resterm * fqterm
				a2 = resterm * resterm
				scale = 1 + a1 + a2
				lastCut, lastRes = cut, res
			}
			val := in[i]
			temp := ym1
			ym1 = scale*val - a1*ym1 - a2*ym2
			ym2 = temp
			out[i] = ym1
		}
	})
}
