package osc

import (
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewPhasor(opts ...ugen.Option) ugen.UGen {
	return New(SamplerFunc(samplePhasor), opts...)
}

func samplePhasor(phase, dPhase, dutyCycle float64) float64 {
	return math.Max(0, math.Min(1, phase/dutyCycle))
}
