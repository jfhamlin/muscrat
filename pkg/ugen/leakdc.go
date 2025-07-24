package ugen

import (
	"context"
	"math"
)

// NewLeakDC creates a DC blocking filter (high-pass filter).
// The coef parameter controls the filter coefficient (0.995 is a good default).
// Higher values (closer to 1) result in lower cutoff frequency.
func NewLeakDC() UGen {
	var (
		x1          float64
		y1          float64
		initialized bool
	)

	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		coef := cfg.InputSamples["coef"]

		// index the last elements to lift the bounds check
		_ = in[len(out)-1]
		_ = coef[len(out)-1]

		// Initialize with the first input sample
		if !initialized {
			x1 = in[0]
			initialized = true
		}

		for i := range out {
			x0 := in[i]
			b1 := coef[i]
			
			// LeakDC formula: y[n] = x[n] - x[n-1] + b1 * y[n-1]
			y1 = x0 - x1 + b1*y1
			
			// Store current input for next iteration
			x1 = x0
			
			// Remove denormal numbers (equivalent to zapgremlins)
			if math.Abs(y1) < 1e-15 {
				y1 = 0
			}
			
			out[i] = y1
		}
	})
}
