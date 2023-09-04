package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewBPF() ugen.UGen {
	// Logic from SuperCollider's BPF

	var freq, bw float64
	var y1, y2, a0, b1, b2 float64

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		ws := cfg.InputSamples["w"]
		bws := cfg.InputSamples["bw"]

		for i := range out {
			if ws[i] != freq || bws[i] != bw {
				pfreq := ws[i] * 2 * math.Pi / float64(cfg.SampleRateHz)
				pbw := bws[i] * pfreq * 0.5

				c := 1 / math.Tan(pbw)
				d := 2 * math.Cos(pfreq)

				a0 = 1 / (1 + c)
				b1 = c * d * a0
				b2 = (1 - c) * a0

				freq = ws[i]
				bw = bws[i]
			}
			y0 := in[i] + b1*y1 + b2*y2
			out[i] = a0 * (y0 - y2)
			y2 = zapgremlins(y1)
			y1 = zapgremlins(y0)
		}
	})
}

func zapgremlins(x float64) float64 {
	absx := math.Abs(x)
	if absx < math.SmallestNonzeroFloat64 || absx > math.MaxFloat64 {
		return 0
	}
	return x
}
