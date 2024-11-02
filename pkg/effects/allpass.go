package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewAllPass(maxDelayTime float64) ugen.UGen {
	delayLine := NewDelayLine(conf.SampleRate, maxDelayTime)

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]

		// For per-sample parameters, we can cache references to the input slices
		delayTimeSamplesInput := cfg.InputSamples["delaytime"]
		decayTimeSamplesInput := cfg.InputSamples["decaytime"]

		lastIdx := len(out) - 1
		_ = in[lastIdx]
		_ = delayTimeSamplesInput[lastIdx]
		_ = decayTimeSamplesInput[lastIdx]

		for i := range out {
			xn := in[i]

			// Get per-sample delayTime
			delayTimeSample := delayTimeSamplesInput[i]
			delayLine.SetDelaySeconds(delayTimeSample)

			// Get per-sample decayTime
			decayTimeSample := decayTimeSamplesInput[i]

			// Compute feedback coefficient g
			delayTime := delayTimeSample
			decayTime := decayTimeSample
			decayAbs := math.Abs(decayTime)
			if decayAbs == 0 {
				decayAbs = 0.0001 // prevent division by zero
			}
			g := math.Exp(-3 * delayTime / decayAbs)
			if decayTime < 0 {
				g = -g
			}

			// Read from delay line
			delayedSample := delayLine.ReadSampleC()

			// All-pass filter equation
			yn := -g*xn + delayedSample

			// Update delay line
			delayLine.WriteSample(xn + g*yn)

			// Output sample
			out[i] = yn
		}
	})
}
