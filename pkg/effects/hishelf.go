package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

const twopi = 2 * math.Pi

func NewHiShelf() ugen.UGen {
	sampleRate := float64(conf.SampleRate)
	var y1, y2, prevFreq, prevRS, prevDB float64
	var a0, a1, a2, b1, b2 float64
	var coefficientsComputed bool

	computeCoefficients := func(freq, rs, db float64) {
		a := math.Pow(10, db*0.025)
		w0 := twopi * freq / sampleRate
		cosw0 := math.Cos(w0)
		sinw0 := math.Sin(w0)
		alpha := sinw0 * 0.5 * math.Sqrt((a+(1/a))*(rs-1)+2)
		i := (a + 1) * cosw0
		j := (a - 1) * cosw0
		k := 2 * math.Sqrt(a) * alpha
		b0rz := 1 / ((a + 1) - j + k)
		a0 = a * ((a + 1) + j + k) * b0rz
		a1 = -2 * a * ((a - 1) + i) * b0rz
		a2 = a * ((a + 1) + j - k) * b0rz
		b1 = -2 * ((a - 1) - i) * b0rz
		b2 = ((a + 1) - j - k) * -b0rz
	}

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		w := cfg.InputSamples["w"]
		rsInput := cfg.InputSamples["rs"]
		dbInput := cfg.InputSamples["db"]

		for i := range out {
			freq := w[i]
			rs := rsInput[i]
			db := dbInput[i]

			if !coefficientsComputed || freq != prevFreq || rs != prevRS || db != prevDB {
				computeCoefficients(freq, rs, db)
				prevFreq, prevRS, prevDB = freq, rs, db
				coefficientsComputed = true
			}

			y0 := in[i] + b1*y1 + b2*y2
			out[i] = a0*y0 + a1*y1 + a2*y2

			y2 = ugen.ZapGremlins(y1)
			y1 = ugen.ZapGremlins(y0)
		}
	})
}
