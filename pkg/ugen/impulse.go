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

	var phaseOffset, freq, phase, phaseIncrement float64
	initialized := false
	lastSync := 1.0

	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		ws := cfg.InputSamples["w"]
		iphases := cfg.InputSamples["iphase"]
		syncs := cfg.InputSamples["sync"]
		if len(syncs) == 0 {
			syncs = Zeros
		}

		_ = ws[len(out)-1]
		_ = syncs[len(out)-1]

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

			if syncs[i] > 0 && lastSync <= 0 {
				phase = 1 // wrap around and trigger a tick
			}
			lastSync = syncs[i]

			if phase >= 1 {
				phase = math.Mod(phase, 1)
				out[i] = 1
			} else {
				out[i] = 0
			}
		}
	})
}
