package osc

import (
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func polyBlep(phase, dPhase float64) float64 {
	if phase < dPhase {
		phase /= dPhase
		return phase + phase - phase*phase - 1.0
	}
	if phase > 1.0-dPhase {
		phase = (phase - 1.0) / dPhase
		return phase*phase + phase + phase + 1.0
	}
	return 0.0
}

func NewSaw(opts ...ugen.Option) ugen.UGen {
	return New(SamplerFunc(sampleSaw), opts...)
}

func sampleSaw(phase, dPhase, dutyCycle float64) float64 {
	phase = dcPhase(phase, dutyCycle)

	result := 2.0*phase - 1.0
	result -= polyBlep(phase, dPhase)
	return result
}

func NewPulse(opts ...ugen.Option) ugen.UGen {
	return New(SamplerFunc(samplePulse), opts...)
}

func samplePulse(phase, dPhase, dutyCycle float64) float64 {
	if dutyCycle >= 1.0 {
		return 1.0
	} else if dutyCycle <= 0.0 {
		return -1.0
	}

	result := -1.0
	if phase >= dutyCycle {
		result = 1.0
	}

	result += polyBlep(math.Mod(phase+dutyCycle, 1), dPhase)
	result -= polyBlep(phase, dPhase)

	return result
}

func NewTri(opts ...ugen.Option) ugen.UGen {
	return New(SamplerFunc(sampleTri), opts...)
}

func sampleTri(phase, dPhase, dutyCycle float64) float64 {
	phase = dcPhase(phase, dutyCycle)

	// TODO: integrate square
	if phase < 0.5 {
		return 4.0*phase - 1.0
	}
	return 3.0 - 4.0*phase
}

func dcPhase(phase, dc float64) float64 {
	return math.Max(0, math.Min(1, phase/dc))
}
