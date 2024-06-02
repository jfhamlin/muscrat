package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewRHPF() ugen.UGen {
	sampleRate := float64(conf.SampleRate)
	radiansPerSample := twopi / sampleRate
	var y1, y2, prevFreq, prevReson float64
	var a0, b1, b2 float64
	var coefficientsComputed bool

	computeCoefficients := func(freq, reson float64) {
		qres := math.Max(0.001, reson)
		pfreq := freq * radiansPerSample

		D := math.Tan(pfreq * qres * 0.5)
		C := (1.0 - D) / (1.0 + D)
		cosf := math.Cos(pfreq)

		b1 = (1.0 + C) * cosf
		b2 = -C
		a0 = (1.0 + C + b1) * 0.25
	}

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		freq := cfg.InputSamples["freq"]
		reson := cfg.InputSamples["reson"]

		_ = in[len(out)-1]
		_ = freq[len(out)-1]
		_ = reson[len(out)-1]

		for i := range out {
			if !coefficientsComputed || freq[i] != prevFreq || reson[i] != prevReson {
				computeCoefficients(freq[i], reson[i])
				prevFreq, prevReson = freq[i], reson[i]
				coefficientsComputed = true
			}

			y0 := a0*in[i] + b1*y1 + b2*y2
			out[i] = y0 - 2.0*y1 + y2

			y2 = ugen.ZapGremlins(y1)
			y1 = ugen.ZapGremlins(y0)
		}
	})
}
