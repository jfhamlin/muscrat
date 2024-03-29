package osc

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

type (
	Sampler interface {
		// Sample returns the sample value for the given phase and phase
		// delta per sample.
		Sample(phase, dPhase, dutyCycle float64) float64
	}

	SamplerFunc func(phase, dPhase, dutyCycle float64) float64

	Osc struct {
		options ugen.Options
		sampler Sampler

		initialized bool

		initialPhase float64
		lastPhase    float64
		lastSync     float64
	}
)

func New(s Sampler, opts ...ugen.Option) *Osc {
	options := ugen.DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return &Osc{
		options: options,
		sampler: s,
	}
}

// TODO:
// auto-gen variants of the inner loop for different input combinations
//
// phase overrides frequency
// phase overrides iphase
//
// combinations:
//
// ws: [syncs, dcs, iphases] => 8 variants
// phases: [dcs] => 2 variants
//
// 16 + 8 = 24 variants

func (o *Osc) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	ws := cfg.InputSamples["w"]
	// TODO: generic test that checks that manually oscillating the
	// phase at a given frequency produces the same output as just
	// using the frequency input.
	phases := cfg.InputSamples["phase"]
	syncs := cfg.InputSamples["sync"]
	dcs := cfg.InputSamples["dc"]
	iphases := cfg.InputSamples["iphase"] // todo: handle changing iphase

	if !o.initialized {
		o.initialized = true
		if len(iphases) > 0 {
			o.initialPhase = iphases[0]
			o.lastPhase = o.initialPhase
		}
	}

	phase := o.lastPhase
	sampler := o.sampler
	lastSync := o.lastSync

	// TODO: pull all the conditional logic out of the loop

	for i := range out {
		dc := o.options.DefaultDutyCycle
		if len(dcs) > 0 {
			dc = dcs[i]
		}
		var dPhase float64
		w := 440.0 // default frequency
		if len(ws) > 0 {
			w = ws[i]
			dPhase = w / float64(cfg.SampleRateHz)
		}
		if len(phases) > 0 {
			dPhase = (phases[i] - phase)
			w = dPhase * float64(cfg.SampleRateHz)
			phase = phases[i]
			// keep in [0, 1)
			phase -= math.Floor(phase)
		}

		out[i] = sampler.Sample(phase, dPhase, dc)

		if len(phases) == 0 {
			phase += w / float64(cfg.SampleRateHz)
			phase -= math.Floor(phase)
		}

		// sync on the falling edge of the sync input if present
		if len(syncs) > 0 {
			if syncs[i] < lastSync {
				phase = 0.0
			}
			lastSync = syncs[i]
		}
	}

	o.lastPhase = phase
	o.lastSync = lastSync
}

func (f SamplerFunc) Sample(phase, dPhase, dutyCycle float64) float64 {
	return f(phase, dPhase, dutyCycle)
}
