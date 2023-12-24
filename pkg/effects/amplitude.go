package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

const (
	sampleRate = 44100 // TODO: move to central, configurable location
)

var (
	log1 = math.Log(0.1)
)

func NewAmplitude(attackTime, releaseTime float64, opts ...ugen.Option) ugen.UGen {
	o := ugen.DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	add := o.Add
	mul := o.Mul

	clampCoef := 0.0
	if attackTime != 0 {
		math.Exp(log1 / (attackTime * sampleRate))
	}
	relaxCoef := 0.0
	if releaseTime != 0 {
		math.Exp(log1 / (releaseTime * sampleRate))
	}

	prevIn := 0.0

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		// code ported from Supercollider's Amplitdue ugen
		// https://doc.sccode.org/Classes/Amplitude.html

		in := cfg.InputSamples["in"]
		for i := range out {
			val := math.Abs(in[i])
			if val < prevIn {
				val = val + (prevIn-val)*relaxCoef
			} else {
				val = val + (prevIn-val)*clampCoef
			}
			prevIn = val
			out[i] = mul*val + add
		}
	})
}
