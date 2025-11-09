package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewLPF() ugen.UGen {
	sampleRate := float64(conf.SampleRate)
	radiansPerSample := twopi / sampleRate
	var y1, y2, prevFreq float64
	var a0, b1, b2 float64
	var coefficientsComputed bool

	computeCoefficients := func(freq float64) {
		// SuperCollider LPF uses Butterworth filter design
		pfreq := freq * radiansPerSample * 0.5

		C := 1.0 / math.Tan(pfreq)
		C2 := C * C
		sqrt2C := C * math.Sqrt2

		a0 = 1.0 / (1.0 + sqrt2C + C2)
		b1 = -2.0 * (1.0 - C2) * a0
		b2 = -(1.0 - sqrt2C + C2) * a0
	}

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		freq := cfg.InputSamples["freq"]

		_ = in[len(out)-1]
		_ = freq[len(out)-1]

		for i := range out {
			if !coefficientsComputed || freq[i] != prevFreq {
				computeCoefficients(freq[i])
				prevFreq = freq[i]
				coefficientsComputed = true
			}

			y0 := in[i] + b1*y1 + b2*y2
			out[i] = a0 * (y0 + 2.0*y1 + y2)

			y2 = ugen.ZapGremlins(y1)
			y1 = ugen.ZapGremlins(y0)
		}
	})
}
