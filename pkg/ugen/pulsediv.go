package ugen

import (
	"context"
	"math"
)

func NewPulseDiv(start float64) UGen {
	count := int(math.Floor(start + 0.5))
	lastTrig := 0.0
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		trig := cfg.InputSamples["trigger"]
		div := cfg.InputSamples["div"]

		_ = trig[len(out)-1]
		_ = div[len(out)-1]

		for i := range out {
			if trig[i] > 0 && lastTrig <= 0 {
				count++
				if count >= int(div[i]) {
					out[i] = 1
					count = 0
				} else {
					out[i] = 0
				}
			} else {
				out[i] = 0
			}
			lastTrig = trig[i]
		}
	})
}
