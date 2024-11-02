package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewDelay(maxDelay float64, opts ...ugen.Option) ugen.UGen {
	o := ugen.DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	delayLine := NewDelayLine(conf.SampleRate, maxDelay)
	var delaySeconds float64

	readSample := delayLine.ReadSampleN
	switch o.Interp {
	case ugen.InterpNone:
		readSample = delayLine.ReadSampleN
	case ugen.InterpLinear:
		readSample = delayLine.ReadSampleL
	case ugen.InterpCubic:
		readSample = delayLine.ReadSampleC
	default:
		panic("unknown interpolation type")
	}

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		delays := cfg.InputSamples["delay"]

		for i := range out {
			newDelaySeconds := math.Max(0, delays[i])
			newDelaySeconds = math.Min(newDelaySeconds, maxDelay)
			if newDelaySeconds != delaySeconds {
				delayLine.SetDelaySeconds(newDelaySeconds)
				delaySeconds = newDelaySeconds
			}

			delayLine.WriteSample(in[i])
			out[i] = readSample()
		}
	})
}
