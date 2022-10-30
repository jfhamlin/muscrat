package wavtabs

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
)

type genOpts struct {
	defaultDutyCycle float64
}

// GeneratorOption is an option for the Generator.
type GeneratorOption func(*genOpts)

// WithDefaultDutyCycle sets the default duty cycle for the wavetable.
func WithDefaultDutyCycle(dc float64) GeneratorOption {
	return func(opts *genOpts) {
		opts.defaultDutyCycle = dc
	}
}

// Generator is a generator that generates a wavetable.
func Generator(wavtab Table, opts ...GeneratorOption) generator.SampleGenerator {
	options := genOpts{
		defaultDutyCycle: 1,
	}
	for _, opt := range opts {
		opt(&options)
	}

	phase := 0.0
	lastSync := 0.0

	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		ws := cfg.InputSamples["w"]
		phases := cfg.InputSamples["phase"]
		syncs := cfg.InputSamples["sync"]
		dcs := cfg.InputSamples["dc"]

		res := make([]float64, n)
		for i := 0; i < n; i++ {
			if i < len(phases) {
				phase = phases[i]
			}
			dc := options.defaultDutyCycle
			if i < len(dcs) {
				dc = dcs[i]
			}
			switch dc {
			case 0:
				res[i] = wavtab[0]
			case 1:
				res[i] = wavtab.Lerp(phase)
			default:
				t := (phase - math.Floor(phase)) / dc
				if t > 1 {
					res[i] = wavtab[len(wavtab)-1]
				} else {
					res[i] = wavtab.Lerp(t)
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
		return res
	})
}
