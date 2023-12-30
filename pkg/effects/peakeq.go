package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewPeakEQ(opts ...ugen.Option) ugen.UGen {
	o := ugen.DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	add := o.Add
	mul := o.Mul

	const twopi = 2 * math.Pi

	sampleRate := float64(conf.SampleRate)
	var y1, y2, prevFreq, prevRQ, prevDB float64
	var a0, a1, a2, b1, b2 float64
	var coefficientsComputed bool

	computeCoefficients := func(freq, rq, db float64) {
		a := math.Pow(10, db*0.025)
		w0 := twopi * freq / sampleRate
		alpha := math.Sin(w0) * 0.5 * rq
		b0rz := 1 / (1 + (alpha / a))
		b1 = 2 * b0rz * math.Cos(w0)
		a0 = (1 + (alpha * a)) * b0rz
		a1 = -b1
		a2 = (1 - (alpha * a)) * b0rz
		b2 = (1 - (alpha / a)) * -b0rz
	}

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		w := cfg.InputSamples["w"]
		rqInput := cfg.InputSamples["rq"]
		dbInput := cfg.InputSamples["db"]

		for i := range out {
			freq := w[i]
			rq := rqInput[i]
			db := dbInput[i]

			if !coefficientsComputed || freq != prevFreq || rq != prevRQ || db != prevDB {
				computeCoefficients(freq, rq, db)
				prevFreq, prevRQ, prevDB = freq, rq, db
				coefficientsComputed = true
			}

			y0 := in[i] + b1*y1 + b2*y2
			out[i] = mul*(a0*y0+a1*y1+a2*y2) + add

			y2 = ugen.ZapGremlins(y1)
			y1 = ugen.ZapGremlins(y0)
		}
	})
}
