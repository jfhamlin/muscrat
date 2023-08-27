package wavtabs

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

type genOpts struct {
	defaultDutyCycle float64
	multiply         float64
	add              float64
}

// GeneratorOption is an option for the Generator.
type GeneratorOption func(*genOpts)

// WithDefaultDutyCycle sets the default duty cycle for the wavetable.
func WithDefaultDutyCycle(dc float64) GeneratorOption {
	return func(opts *genOpts) {
		opts.defaultDutyCycle = dc
	}
}

func WithMultiply(m float64) GeneratorOption {
	return func(opts *genOpts) {
		opts.multiply = m
	}
}

func WithAdd(a float64) GeneratorOption {
	return func(opts *genOpts) {
		opts.add = a
	}
}

// Generator is a generator that generates a wavetable.
func Generator(wavtabIn Table, opts ...GeneratorOption) ugen.UGen {
	options := genOpts{
		defaultDutyCycle: 1,
		multiply:         1,
	}
	for _, opt := range opts {
		opt(&options)
	}

	wavtab := make(Table, len(wavtabIn))
	for i, v := range wavtabIn {
		wavtab[i] = v*options.multiply + options.add
	}

	phase := 0.0
	lastSync := 0.0

	initialised := false

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		ws := cfg.InputSamples["w"]
		phases := cfg.InputSamples["phase"]
		syncs := cfg.InputSamples["sync"]
		dcs := cfg.InputSamples["dc"]
		// TODO: the common case is to set this once at the start. There
		// are some semantics to figure out here, but it would be nice to
		// be able to set this once at the start and then have it apply
		// to all the samples.
		iphases := cfg.InputSamples["iphase"]
		if !initialised {
			initialised = true

			if len(iphases) > 0 {
				phase = iphases[0]
			}
		}

		// TODO: band-limited interpolation

		for i := range out {
			if i < len(phases) {
				phase = phases[i]
			}
			dc := options.defaultDutyCycle
			if i < len(dcs) {
				dc = dcs[i]
			}
			switch dc {
			case 0:
				out[i] = wavtab[0]
			case 1:
				out[i] = wavtab.Lerp(phase)
			default:
				t := (phase - math.Floor(phase)) / dc
				if t > 1 {
					out[i] = wavtab[len(wavtab)-1]
				} else {
					out[i] = wavtab.Lerp(t)
				}
			}

			w := 440.0 // default frequency
			if i < len(ws) {
				w = ws[i]
			}
			phase += w / float64(cfg.SampleRateHz)
			// keep phase in [0, 1)
			phase -= math.Floor(phase)

			// sync on the falling edge of the sync input if present
			if i < len(syncs) {
				if syncs[i] < lastSync {
					phase = 0.0
				}
				lastSync = syncs[i]
			}
		}
	})
}
