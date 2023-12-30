package ugen

import (
	"context"
	"math"
)

func NewImpulse(opts ...Option) UGen {
	o := DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	add := o.Add
	mul := o.Mul

	var phaseOffset, freq, phase, phaseIncrement float64
	initialized := false

	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
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

		for i := range out {
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
				out[i] = mul*1 + add
			} else {
				out[i] = add
			}
		}
	})
}
