package sampler

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewSampler(buf []float64, loop bool) ugen.UGen {
	if len(buf) == 0 {
		return ugen.NewConstant(0)
	}

	sampleLen := float64(len(buf))
	index := 0.0
	lastGate := false
	stopped := true
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		gate := cfg.InputSamples["trigger"]
		rates := cfg.InputSamples["rate"]
		if len(gate) == 0 {
			// always play if no trigger
			stopped = false
			lastGate = true
		}

		for i := range out {
			if !lastGate && gate[i] > 0 {
				stopped = false
				index = 0
			}
			if len(gate) > 0 {
				lastGate = gate[i] > 0
			}
			rate := 1.0
			if len(rates) > 0 {
				rate = rates[i]
				if rate < 0 {
					rate = 0
				}
			}

			if stopped {
				continue
			}

			if index >= sampleLen {
				out[i] = 0
			} else {
				// cubic interpolation
				sampleIndexF, frac := math.Modf(index)
				sampleIndex := int(sampleIndexF)

				s0 := buf[sampleIndex]
				var s1, s2, s3 float64
				sample := s0
				if sampleIndex+1 < len(buf) {
					s1 = buf[sampleIndex+1]
					sample += (s1 - s0) * frac
				}
				if sampleIndex+2 < len(buf) {
					s2 = buf[sampleIndex+2]
					sample += (0.5*s2 - s1 + 0.5*s0) * frac * frac
				}
				if sampleIndex+3 < len(buf) {
					s3 = buf[sampleIndex+3]
					sample += (-0.5*s3 + 1.5*s2 - 1.5*s1 + 0.5*s0) * frac * frac * frac
				}

				out[i] = sample
			}

			index += rate
			if index >= sampleLen {
				index = 0
				if !loop || len(gate) > 0 && !lastGate {
					stopped = true
				}
			}
		}
	})
}
