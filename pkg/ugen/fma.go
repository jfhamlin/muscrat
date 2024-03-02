package ugen

import "context"

// NewFMA creates a new ugen for a fused multiply + add operation with
// dynamic multiplicand and summand.
func NewFMA() UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		mul := cfg.InputSamples["mul"]
		add := cfg.InputSamples["add"]
		// index the last element of in, mul, and add to lift the bounds check
		_ = in[len(out)-1]
		_ = mul[len(out)-1]
		_ = add[len(out)-1]
		for i := range out {
			out[i] = mul[i]*in[i] + add[i]
		}
	})
}

// NewFMAStatic creates a new ugen for a fused multiply + add
// operation with static multiplicand and summand.
func NewFMAStatic(mul, add float64) UGen {
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		// index the last element of in to lift the bounds check
		_ = in[len(out)-1]
		for i := range out {
			out[i] = mul*in[i] + add
		}
	})
}
