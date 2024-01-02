package osc

import "github.com/jfhamlin/muscrat/pkg/ugen"

// NewLFSaw returns a new low-frequency (non-band-limited, unipolar) sawtooth
// oscillator.
func NewLFSaw(opts ...ugen.Option) ugen.UGen {
	return New(SamplerFunc(func(phase, dPhase, dutyCycle float64) float64 {
		phase = dcPhase(phase, dutyCycle)
		return phase
	}), opts...)
}

func NewLFPulse(opts ...ugen.Option) ugen.UGen {
	return New(SamplerFunc(func(phase, dPhase, dutyCycle float64) float64 {
		if phase >= dutyCycle {
			return 0
		}
		return 1.0
	}), opts...)
}
