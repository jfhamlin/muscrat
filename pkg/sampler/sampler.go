package sampler

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewSampler(buf []float64) ugen.UGen {
	if len(buf) == 0 {
		return ugen.NewConstant(0)
	}

	sampleLen := float64(len(buf))
	index := 0.0
	lastTrig := false
	stopped := true
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		trigs := cfg.InputSamples["trigger"]
		rates := cfg.InputSamples["rate"]
		loops := cfg.InputSamples["loop"]
		startIndex := cfg.InputSamples["start-pos"]
		endIndex := cfg.InputSamples["end-pos"]

		lastIdx := len(out) - 1
		_ = trigs[lastIdx]
		_ = rates[lastIdx]
		_ = loops[lastIdx]
		_ = startIndex[lastIdx]
		_ = endIndex[lastIdx]

		for i := range out {
			if !lastTrig && trigs[i] > 0 {
				stopped = false
				index = startIndex[i]
			}
			lastTrig = trigs[i] > 0

			rate := rates[i]

			end := endIndex[i]
			if end == 0 {
				end = 1
			} else if end > sampleLen {
				end = sampleLen
			}

			if stopped {
				continue
			}

			if index >= end {
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
			if index < 0 {
				index = float64(len(buf) - 1)
			}
			if index >= end {
				index = startIndex[i]
				if loops[i] <= 0 {
					stopped = true
				}
			}
		}
	})
}
