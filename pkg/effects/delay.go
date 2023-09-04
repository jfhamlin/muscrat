package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewDelay(maxDelay float64, opts ...ugen.Option) ugen.UGen {
	o := ugen.DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	add := o.Add
	mul := o.Mul

	var mask int
	var buf []float64
	var writePos int

	var interp func(idx int, frac float64) float64
	switch o.Interp {
	case ugen.InterpNone:
		interp = func(idx int, frac float64) float64 {
			return buf[idx&mask]
		}
	case ugen.InterpLinear:
		interp = func(idx int, frac float64) float64 {
			return ugen.LinInterp(frac, buf[idx&mask], buf[(idx-1)&mask])
		}
	case ugen.InterpCubic:
		interp = func(idx int, frac float64) float64 {
			x0 := buf[(idx+1)&mask]
			x1 := buf[idx&mask]
			x2 := buf[(idx-1)&mask]
			x3 := buf[(idx-2)&mask]
			return ugen.CubInterp(frac, x0, x1, x2, x3)
		}
	default:
		panic("unknown interpolation type")
	}

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		if buf == nil {
			sz := ugen.NextPowerOf2(int(math.Ceil(maxDelay*float64(cfg.SampleRateHz) + 1)))
			mask = sz - 1
			buf = make([]float64, sz)
		}

		in := cfg.InputSamples["in"]
		delays := cfg.InputSamples["delay"]

		for i := range out {
			delaySeconds := delays[i]
			if delaySeconds > maxDelay {
				delaySeconds = maxDelay
			}
			if delaySeconds < 0 {
				delaySeconds = 0
			}
			buf[writePos&mask] = in[i]

			delaySamples := delaySeconds * float64(cfg.SampleRateHz)
			delaySamplesInt, delaySamplesFrac := math.Modf(delaySamples)

			readPos := writePos - int(delaySamplesInt)
			out[i] = mul*interp(readPos, delaySamplesFrac) + add

			writePos++
		}
	})
}
