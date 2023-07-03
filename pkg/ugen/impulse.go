package ugen

import (
	"context"
	"math"
)

func NewImpulse() UGen {
	var phaseOffset, freq, phase, phaseIncrement float64
	initialized := false

	return SampleGeneratorFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		ws := cfg.InputSamples["w"]
		iphases := cfg.InputSamples["iphase"]

		if !initialized {
			freq = math.Max(ws[0], 0)
			if len(iphases) > 0 {
				phaseOffset = iphases[0]
				phase = math.Mod(phaseOffset, 1)
				if phase < 0 {
					phase = 1 + phase
				}
			}
			if phase == 0 {
				phase = 1 // emit a sample on the first iteration
			}
			phaseIncrement = freq / float64(cfg.SampleRateHz)
			initialized = true
		}

		res := make([]float64, n)
		for i := 0; i < n; i++ {
			if ws[i] != freq {
				freq = math.Max(ws[i], 0)
				phaseIncrement = freq / float64(cfg.SampleRateHz)
			}
			if len(iphases) > 0 && iphases[i] != phaseOffset {
				correction := iphases[i] - phaseOffset
				phaseOffset = iphases[i]
				phase = math.Mod(phase+correction, 1)
				if phase < 0 {
					phase = 1 + phase
				}
			}
			phase += phaseIncrement
			if phase >= 1 {
				phase = math.Mod(phase, 1)
				res[i] = 1
			}
		}
		return res
	})
}
