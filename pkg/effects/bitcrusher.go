package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewBitcrusher() ugen.UGen {
	const (
		defaultDepth = 24
		defaultRate  = 44100
	)
	var defaultRates, defaultDepths []float64

	var count, lastOut float64

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		in := cfg.InputSamples["in"]
		rate := cfg.InputSamples["rate"]
		bits := cfg.InputSamples["bits"]

		if len(rate) == 0 {
			if len(defaultRates) < n {
				for i := len(defaultRates); i < n; i++ {
					defaultRates = append(defaultRates, defaultRate)
				}
			}
			rate = defaultRates
		}
		if len(bits) == 0 {
			if len(defaultDepths) < n {
				for i := len(defaultDepths); i < n; i++ {
					defaultDepths = append(defaultDepths, defaultDepth)
				}
			}
			bits = defaultDepths
		}

		res := make([]float64, n)
		for i := 0; i < n; i++ {
			var step, stepr, ratio float64
			if bits[i] >= 31 || bits[i] < 1 {
				step = 0
				stepr = 1
			} else {
				step = math.Pow(0.5, bits[i]-0.999)
				stepr = 1 / step
			}
			if rate[i] >= float64(cfg.SampleRateHz) {
				ratio = 1
			} else {
				ratio = rate[i] / float64(cfg.SampleRateHz)
			}

			count += ratio
			if count >= 1 {
				count -= 1
				offset := 1.0
				if in[i] < 0 {
					offset = -1.0
				}
				_, delta := math.Modf((in[i] - offset*step*0.5) * stepr)
				delta *= step
				lastOut = in[i] - delta
			}
			res[i] = lastOut
		}
		return res
	})
}
