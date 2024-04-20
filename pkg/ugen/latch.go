package ugen

import "context"

func NewLatch() UGen {
	lastTrig := 0.0
	val := 0.0
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		trig := cfg.InputSamples["trigger"]
		// index the last element of in to lift the bounds check
		_ = in[len(out)-1]
		_ = trig[len(out)-1]
		for i := range out {
			if trig[i] > 0 && lastTrig <= 0 {
				val = in[i]
			}
			out[i] = val
			lastTrig = trig[i]
		}
	})
}
