package osc

import "github.com/jfhamlin/muscrat/pkg/ugen"

type (
	phasor struct{}
)

func NewPhasor(opts ...ugen.Option) ugen.UGen {
	return New(phasor{}, opts...)
}

func (p phasor) Sample(phase, dPhase float64) float64 {
	return phase
}
