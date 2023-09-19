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

type (
	saw struct{}

	square struct{}

	tri struct{}
)

func NewSaw(opts ...ugen.Option) ugen.UGen {
	return New(saw{}, opts...)
}

func (s saw) Sample(phase, dPhase float64) float64 {
	result := 2.0*phase - 1.0
	result -= polyBlep(phase, dPhase)
	return result
}

func NewSquare(opts ...ugen.Option) ugen.UGen {
	return New(square{}, opts...)
}

func (s square) Sample(phase, dPhase float64) float64 {
	result := -1.0
	if phase >= 0.5 {
		result = 1.0
	}

	result += polyBlep(math.Mod(phase+0.5, 1), dPhase)
	result -= polyBlep(phase, dPhase)

	return result
}

func NewTri(opts ...ugen.Option) ugen.UGen {
	return New(tri{}, opts...)
}

func (t tri) Sample(phase, dPhase float64) float64 {
	// TODO: integrate square
	if phase < 0.5 {
		return 4.0*phase - 1.0
	}
	return 3.0 - 4.0*phase
}
