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

		initialPhase    float64
		phase           float64
		lastSamplePhase float64
		lastSync        float64
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
// combinations:
//
// ws: [syncs, dcs, iphases] => 8 variants
// phases: [dcs] => 2 variants
//
// 16 + 8 = 24 variants

func (o *Osc) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	ws := cfg.InputSamples["w"]
	phases := cfg.InputSamples["phase"]
	syncs := cfg.InputSamples["sync"]
	dcs := cfg.InputSamples["dc"]
	iphases := cfg.InputSamples["iphase"]

	if !o.initialized {
		o.initialized = true
		initialW := 440.0
		if len(ws) > 0 {
			initialW = ws[0]
		}
		if len(iphases) > 0 {
			o.initialPhase = iphases[0]
			o.phase = o.initialPhase - (initialW / float64(cfg.SampleRateHz))
			o.lastSamplePhase = o.initialPhase - (initialW / float64(cfg.SampleRateHz))
		}
		if len(phases) > 0 {
			o.lastSamplePhase += phases[0]
		}
		mod1(&o.initialPhase)
		mod1(&o.lastSamplePhase)
		mod1(&o.phase)
	}

	phase := o.phase
	lastSamplePhase := o.lastSamplePhase
	sampler := o.sampler
	lastSync := o.lastSync

	// TODO: pull all the conditional logic out of the loop

	for i := range out {
		dc := o.options.DefaultDutyCycle
		if len(dcs) > 0 {
			dc = dcs[i]
		}
		w := 440.0 // default frequency
		if len(ws) > 0 {
			w = ws[i]
		}

		samplePhase := phase
		if len(phases) > 0 {
			// phase is an offset in [0, 1)
			samplePhase += phases[i]
			mod1(&samplePhase)
		}

		dPhase := samplePhase - lastSamplePhase
		lastSamplePhase = samplePhase
		out[i] = sampler.Sample(samplePhase, dPhase, dc)

		phase += w / float64(cfg.SampleRateHz)
		mod1(&phase)

		// sync on the falling edge of the sync input if present
		if len(syncs) > 0 {
			if syncs[i] < lastSync {
				phase = 0.0
			}
			lastSync = syncs[i]
		}
	}

	o.phase = phase
	o.lastSamplePhase = lastSamplePhase
	o.lastSync = lastSync
}

func (f SamplerFunc) Sample(phase, dPhase, dutyCycle float64) float64 {
	return f(phase, dPhase, dutyCycle)
}

func mod1(x *float64) {
	*x -= math.Floor(*x)
}
