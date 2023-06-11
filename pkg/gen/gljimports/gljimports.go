// GENERATED FILE. DO NOT EDIT.
package gljimports

import (
	github_com_jfhamlin_muscrat_pkg_ugen "github.com/jfhamlin/muscrat/pkg/ugen"
	github_com_jfhamlin_muscrat_pkg_wavtabs "github.com/jfhamlin/muscrat/pkg/wavtabs"
	github_com_jfhamlin_muscrat_pkg_stochastic "github.com/jfhamlin/muscrat/pkg/stochastic"
	github_com_jfhamlin_muscrat_pkg_effects "github.com/jfhamlin/muscrat/pkg/effects"
	github_com_jfhamlin_muscrat_pkg_mod "github.com/jfhamlin/muscrat/pkg/mod"
	github_com_jfhamlin_muscrat_pkg_sampler "github.com/jfhamlin/muscrat/pkg/sampler"
	github_com_jfhamlin_muscrat_pkg_midi "github.com/jfhamlin/muscrat/pkg/midi"
	github_com_jfhamlin_muscrat_pkg_aio "github.com/jfhamlin/muscrat/pkg/aio"
	github_com_jfhamlin_muscrat_pkg_graph "github.com/jfhamlin/muscrat/pkg/graph"
	github_com_jfhamlin_muscrat_pkg_pattern "github.com/jfhamlin/muscrat/pkg/pattern"
	github_com_jfhamlin_freeverb_go "github.com/jfhamlin/freeverb-go"
	github_com_jfhamlin_muscrat_pkg_slice "github.com/jfhamlin/muscrat/pkg/slice"
	"reflect"
)

func RegisterImports(_register func(string, interface{})) {
	// package github.com/jfhamlin/muscrat/pkg/ugen
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/ugen.CollectIndexedInputs", github_com_jfhamlin_muscrat_pkg_ugen.CollectIndexedInputs)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewConstant", github_com_jfhamlin_muscrat_pkg_ugen.NewConstant)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewFreqRatio", github_com_jfhamlin_muscrat_pkg_ugen.NewFreqRatio)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewProduct", github_com_jfhamlin_muscrat_pkg_ugen.NewProduct)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewSum", github_com_jfhamlin_muscrat_pkg_ugen.NewSum)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.SampleConfig", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SampleConfig)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.SampleGenerator", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SampleGenerator)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.SampleGeneratorFunc", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SampleGeneratorFunc)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.SimpleSampleGeneratorFunc", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SimpleSampleGeneratorFunc)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Starter", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Starter)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Stopper", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Stopper)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.UGen", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.UGen)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.UGenFunc", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.UGenFunc)(nil)).Elem())

	// package github.com/jfhamlin/muscrat/pkg/wavtabs
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.DefaultResolution", github_com_jfhamlin_muscrat_pkg_wavtabs.DefaultResolution)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.Generator", github_com_jfhamlin_muscrat_pkg_wavtabs.Generator)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.GeneratorOption", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_wavtabs.GeneratorOption)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.Phasor", github_com_jfhamlin_muscrat_pkg_wavtabs.Phasor)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.Pulse", github_com_jfhamlin_muscrat_pkg_wavtabs.Pulse)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.Saw", github_com_jfhamlin_muscrat_pkg_wavtabs.Saw)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.Sin", github_com_jfhamlin_muscrat_pkg_wavtabs.Sin)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.Table", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_wavtabs.Table)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.Tri", github_com_jfhamlin_muscrat_pkg_wavtabs.Tri)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.WithAdd", github_com_jfhamlin_muscrat_pkg_wavtabs.WithAdd)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.WithDefaultDutyCycle", github_com_jfhamlin_muscrat_pkg_wavtabs.WithDefaultDutyCycle)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.WithMultiply", github_com_jfhamlin_muscrat_pkg_wavtabs.WithMultiply)

	// package github.com/jfhamlin/muscrat/pkg/stochastic
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.NewNoise", github_com_jfhamlin_muscrat_pkg_stochastic.NewNoise)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.NewNoiseQuad", github_com_jfhamlin_muscrat_pkg_stochastic.NewNoiseQuad)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.NewPinkNoise", github_com_jfhamlin_muscrat_pkg_stochastic.NewPinkNoise)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.Option", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_stochastic.Option)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.PinkNoise", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_stochastic.PinkNoise)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.WithAdd", github_com_jfhamlin_muscrat_pkg_stochastic.WithAdd)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.WithMul", github_com_jfhamlin_muscrat_pkg_stochastic.WithMul)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.WithRand", github_com_jfhamlin_muscrat_pkg_stochastic.WithRand)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.WithSeed", github_com_jfhamlin_muscrat_pkg_stochastic.WithSeed)

	// package github.com/jfhamlin/muscrat/pkg/effects
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewBPF", github_com_jfhamlin_muscrat_pkg_effects.NewBPF)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewDelay", github_com_jfhamlin_muscrat_pkg_effects.NewDelay)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewFreeverb", github_com_jfhamlin_muscrat_pkg_effects.NewFreeverb)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewLowpassFilter", github_com_jfhamlin_muscrat_pkg_effects.NewLowpassFilter)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewWaveFolder", github_com_jfhamlin_muscrat_pkg_effects.NewWaveFolder)
	_register("github.com/jfhamlin/muscrat/pkg/effects.WaveFolder", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_effects.WaveFolder)(nil)).Elem())

	// package github.com/jfhamlin/muscrat/pkg/mod
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/mod.EnvOption", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_mod.EnvOption)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/mod.NewEnvelope", github_com_jfhamlin_muscrat_pkg_mod.NewEnvelope)
	_register("github.com/jfhamlin/muscrat/pkg/mod.WithInterpolation", github_com_jfhamlin_muscrat_pkg_mod.WithInterpolation)
	_register("github.com/jfhamlin/muscrat/pkg/mod.WithReleaseNode", github_com_jfhamlin_muscrat_pkg_mod.WithReleaseNode)

	// package github.com/jfhamlin/muscrat/pkg/sampler
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/sampler.LoadSample", github_com_jfhamlin_muscrat_pkg_sampler.LoadSample)
	_register("github.com/jfhamlin/muscrat/pkg/sampler.NewSampler", github_com_jfhamlin_muscrat_pkg_sampler.NewSampler)

	// package github.com/jfhamlin/muscrat/pkg/midi
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/midi.A0", github_com_jfhamlin_muscrat_pkg_midi.A0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.A1", github_com_jfhamlin_muscrat_pkg_midi.A1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.A2", github_com_jfhamlin_muscrat_pkg_midi.A2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.A3", github_com_jfhamlin_muscrat_pkg_midi.A3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.A4", github_com_jfhamlin_muscrat_pkg_midi.A4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.A5", github_com_jfhamlin_muscrat_pkg_midi.A5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.A6", github_com_jfhamlin_muscrat_pkg_midi.A6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.A7", github_com_jfhamlin_muscrat_pkg_midi.A7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.A8", github_com_jfhamlin_muscrat_pkg_midi.A8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ab0", github_com_jfhamlin_muscrat_pkg_midi.Ab0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ab1", github_com_jfhamlin_muscrat_pkg_midi.Ab1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ab2", github_com_jfhamlin_muscrat_pkg_midi.Ab2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ab3", github_com_jfhamlin_muscrat_pkg_midi.Ab3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ab4", github_com_jfhamlin_muscrat_pkg_midi.Ab4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ab5", github_com_jfhamlin_muscrat_pkg_midi.Ab5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ab6", github_com_jfhamlin_muscrat_pkg_midi.Ab6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ab7", github_com_jfhamlin_muscrat_pkg_midi.Ab7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ab8", github_com_jfhamlin_muscrat_pkg_midi.Ab8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.As0", github_com_jfhamlin_muscrat_pkg_midi.As0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.As1", github_com_jfhamlin_muscrat_pkg_midi.As1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.As2", github_com_jfhamlin_muscrat_pkg_midi.As2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.As3", github_com_jfhamlin_muscrat_pkg_midi.As3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.As4", github_com_jfhamlin_muscrat_pkg_midi.As4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.As5", github_com_jfhamlin_muscrat_pkg_midi.As5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.As6", github_com_jfhamlin_muscrat_pkg_midi.As6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.As7", github_com_jfhamlin_muscrat_pkg_midi.As7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.As8", github_com_jfhamlin_muscrat_pkg_midi.As8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.B0", github_com_jfhamlin_muscrat_pkg_midi.B0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.B1", github_com_jfhamlin_muscrat_pkg_midi.B1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.B2", github_com_jfhamlin_muscrat_pkg_midi.B2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.B3", github_com_jfhamlin_muscrat_pkg_midi.B3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.B4", github_com_jfhamlin_muscrat_pkg_midi.B4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.B5", github_com_jfhamlin_muscrat_pkg_midi.B5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.B6", github_com_jfhamlin_muscrat_pkg_midi.B6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.B7", github_com_jfhamlin_muscrat_pkg_midi.B7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.B8", github_com_jfhamlin_muscrat_pkg_midi.B8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Bb0", github_com_jfhamlin_muscrat_pkg_midi.Bb0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Bb1", github_com_jfhamlin_muscrat_pkg_midi.Bb1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Bb2", github_com_jfhamlin_muscrat_pkg_midi.Bb2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Bb3", github_com_jfhamlin_muscrat_pkg_midi.Bb3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Bb4", github_com_jfhamlin_muscrat_pkg_midi.Bb4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Bb5", github_com_jfhamlin_muscrat_pkg_midi.Bb5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Bb6", github_com_jfhamlin_muscrat_pkg_midi.Bb6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Bb7", github_com_jfhamlin_muscrat_pkg_midi.Bb7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Bb8", github_com_jfhamlin_muscrat_pkg_midi.Bb8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C0", github_com_jfhamlin_muscrat_pkg_midi.C0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C1", github_com_jfhamlin_muscrat_pkg_midi.C1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C2", github_com_jfhamlin_muscrat_pkg_midi.C2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C3", github_com_jfhamlin_muscrat_pkg_midi.C3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C4", github_com_jfhamlin_muscrat_pkg_midi.C4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C5", github_com_jfhamlin_muscrat_pkg_midi.C5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C6", github_com_jfhamlin_muscrat_pkg_midi.C6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C7", github_com_jfhamlin_muscrat_pkg_midi.C7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C8", github_com_jfhamlin_muscrat_pkg_midi.C8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.C9", github_com_jfhamlin_muscrat_pkg_midi.C9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs0", github_com_jfhamlin_muscrat_pkg_midi.Cs0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs1", github_com_jfhamlin_muscrat_pkg_midi.Cs1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs2", github_com_jfhamlin_muscrat_pkg_midi.Cs2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs3", github_com_jfhamlin_muscrat_pkg_midi.Cs3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs4", github_com_jfhamlin_muscrat_pkg_midi.Cs4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs5", github_com_jfhamlin_muscrat_pkg_midi.Cs5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs6", github_com_jfhamlin_muscrat_pkg_midi.Cs6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs7", github_com_jfhamlin_muscrat_pkg_midi.Cs7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs8", github_com_jfhamlin_muscrat_pkg_midi.Cs8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Cs9", github_com_jfhamlin_muscrat_pkg_midi.Cs9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D0", github_com_jfhamlin_muscrat_pkg_midi.D0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D1", github_com_jfhamlin_muscrat_pkg_midi.D1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D2", github_com_jfhamlin_muscrat_pkg_midi.D2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D3", github_com_jfhamlin_muscrat_pkg_midi.D3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D4", github_com_jfhamlin_muscrat_pkg_midi.D4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D5", github_com_jfhamlin_muscrat_pkg_midi.D5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D6", github_com_jfhamlin_muscrat_pkg_midi.D6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D7", github_com_jfhamlin_muscrat_pkg_midi.D7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D8", github_com_jfhamlin_muscrat_pkg_midi.D8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.D9", github_com_jfhamlin_muscrat_pkg_midi.D9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db0", github_com_jfhamlin_muscrat_pkg_midi.Db0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db1", github_com_jfhamlin_muscrat_pkg_midi.Db1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db2", github_com_jfhamlin_muscrat_pkg_midi.Db2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db3", github_com_jfhamlin_muscrat_pkg_midi.Db3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db4", github_com_jfhamlin_muscrat_pkg_midi.Db4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db5", github_com_jfhamlin_muscrat_pkg_midi.Db5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db6", github_com_jfhamlin_muscrat_pkg_midi.Db6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db7", github_com_jfhamlin_muscrat_pkg_midi.Db7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db8", github_com_jfhamlin_muscrat_pkg_midi.Db8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Db9", github_com_jfhamlin_muscrat_pkg_midi.Db9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds0", github_com_jfhamlin_muscrat_pkg_midi.Ds0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds1", github_com_jfhamlin_muscrat_pkg_midi.Ds1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds2", github_com_jfhamlin_muscrat_pkg_midi.Ds2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds3", github_com_jfhamlin_muscrat_pkg_midi.Ds3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds4", github_com_jfhamlin_muscrat_pkg_midi.Ds4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds5", github_com_jfhamlin_muscrat_pkg_midi.Ds5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds6", github_com_jfhamlin_muscrat_pkg_midi.Ds6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds7", github_com_jfhamlin_muscrat_pkg_midi.Ds7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds8", github_com_jfhamlin_muscrat_pkg_midi.Ds8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Ds9", github_com_jfhamlin_muscrat_pkg_midi.Ds9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E0", github_com_jfhamlin_muscrat_pkg_midi.E0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E1", github_com_jfhamlin_muscrat_pkg_midi.E1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E2", github_com_jfhamlin_muscrat_pkg_midi.E2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E3", github_com_jfhamlin_muscrat_pkg_midi.E3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E4", github_com_jfhamlin_muscrat_pkg_midi.E4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E5", github_com_jfhamlin_muscrat_pkg_midi.E5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E6", github_com_jfhamlin_muscrat_pkg_midi.E6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E7", github_com_jfhamlin_muscrat_pkg_midi.E7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E8", github_com_jfhamlin_muscrat_pkg_midi.E8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.E9", github_com_jfhamlin_muscrat_pkg_midi.E9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb0", github_com_jfhamlin_muscrat_pkg_midi.Eb0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb1", github_com_jfhamlin_muscrat_pkg_midi.Eb1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb2", github_com_jfhamlin_muscrat_pkg_midi.Eb2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb3", github_com_jfhamlin_muscrat_pkg_midi.Eb3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb4", github_com_jfhamlin_muscrat_pkg_midi.Eb4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb5", github_com_jfhamlin_muscrat_pkg_midi.Eb5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb6", github_com_jfhamlin_muscrat_pkg_midi.Eb6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb7", github_com_jfhamlin_muscrat_pkg_midi.Eb7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb8", github_com_jfhamlin_muscrat_pkg_midi.Eb8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Eb9", github_com_jfhamlin_muscrat_pkg_midi.Eb9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F0", github_com_jfhamlin_muscrat_pkg_midi.F0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F1", github_com_jfhamlin_muscrat_pkg_midi.F1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F2", github_com_jfhamlin_muscrat_pkg_midi.F2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F3", github_com_jfhamlin_muscrat_pkg_midi.F3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F4", github_com_jfhamlin_muscrat_pkg_midi.F4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F5", github_com_jfhamlin_muscrat_pkg_midi.F5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F6", github_com_jfhamlin_muscrat_pkg_midi.F6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F7", github_com_jfhamlin_muscrat_pkg_midi.F7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F8", github_com_jfhamlin_muscrat_pkg_midi.F8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.F9", github_com_jfhamlin_muscrat_pkg_midi.F9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs0", github_com_jfhamlin_muscrat_pkg_midi.Fs0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs1", github_com_jfhamlin_muscrat_pkg_midi.Fs1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs2", github_com_jfhamlin_muscrat_pkg_midi.Fs2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs3", github_com_jfhamlin_muscrat_pkg_midi.Fs3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs4", github_com_jfhamlin_muscrat_pkg_midi.Fs4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs5", github_com_jfhamlin_muscrat_pkg_midi.Fs5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs6", github_com_jfhamlin_muscrat_pkg_midi.Fs6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs7", github_com_jfhamlin_muscrat_pkg_midi.Fs7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs8", github_com_jfhamlin_muscrat_pkg_midi.Fs8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Fs9", github_com_jfhamlin_muscrat_pkg_midi.Fs9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G0", github_com_jfhamlin_muscrat_pkg_midi.G0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G1", github_com_jfhamlin_muscrat_pkg_midi.G1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G2", github_com_jfhamlin_muscrat_pkg_midi.G2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G3", github_com_jfhamlin_muscrat_pkg_midi.G3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G4", github_com_jfhamlin_muscrat_pkg_midi.G4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G5", github_com_jfhamlin_muscrat_pkg_midi.G5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G6", github_com_jfhamlin_muscrat_pkg_midi.G6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G7", github_com_jfhamlin_muscrat_pkg_midi.G7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G8", github_com_jfhamlin_muscrat_pkg_midi.G8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.G9", github_com_jfhamlin_muscrat_pkg_midi.G9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb0", github_com_jfhamlin_muscrat_pkg_midi.Gb0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb1", github_com_jfhamlin_muscrat_pkg_midi.Gb1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb2", github_com_jfhamlin_muscrat_pkg_midi.Gb2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb3", github_com_jfhamlin_muscrat_pkg_midi.Gb3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb4", github_com_jfhamlin_muscrat_pkg_midi.Gb4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb5", github_com_jfhamlin_muscrat_pkg_midi.Gb5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb6", github_com_jfhamlin_muscrat_pkg_midi.Gb6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb7", github_com_jfhamlin_muscrat_pkg_midi.Gb7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb8", github_com_jfhamlin_muscrat_pkg_midi.Gb8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gb9", github_com_jfhamlin_muscrat_pkg_midi.Gb9)
	_register("github.com/jfhamlin/muscrat/pkg/midi.GetNote", github_com_jfhamlin_muscrat_pkg_midi.GetNote)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gs0", github_com_jfhamlin_muscrat_pkg_midi.Gs0)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gs1", github_com_jfhamlin_muscrat_pkg_midi.Gs1)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gs2", github_com_jfhamlin_muscrat_pkg_midi.Gs2)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gs3", github_com_jfhamlin_muscrat_pkg_midi.Gs3)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gs4", github_com_jfhamlin_muscrat_pkg_midi.Gs4)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gs5", github_com_jfhamlin_muscrat_pkg_midi.Gs5)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gs6", github_com_jfhamlin_muscrat_pkg_midi.Gs6)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gs7", github_com_jfhamlin_muscrat_pkg_midi.Gs7)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Gs8", github_com_jfhamlin_muscrat_pkg_midi.Gs8)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Names", github_com_jfhamlin_muscrat_pkg_midi.Names)
	_register("github.com/jfhamlin/muscrat/pkg/midi.Note", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_midi.Note)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/midi.Notes", github_com_jfhamlin_muscrat_pkg_midi.Notes)

	// package github.com/jfhamlin/muscrat/pkg/aio
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/aio.InputDevice", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.InputDevice)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.MIDIDevice", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.MIDIDevice)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.MIDIDeviceOption", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.MIDIDeviceOption)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewInputDevice", github_com_jfhamlin_muscrat_pkg_aio.NewInputDevice)
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewMIDIDevice", github_com_jfhamlin_muscrat_pkg_aio.NewMIDIDevice)
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewQwertyMIDI", github_com_jfhamlin_muscrat_pkg_aio.NewQwertyMIDI)
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewSoftwareKeyboard", github_com_jfhamlin_muscrat_pkg_aio.NewSoftwareKeyboard)
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewWavOut", github_com_jfhamlin_muscrat_pkg_aio.NewWavOut)
	_register("github.com/jfhamlin/muscrat/pkg/aio.SoftwareKeyboard", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.SoftwareKeyboard)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.StdinChan", github_com_jfhamlin_muscrat_pkg_aio.StdinChan)
	_register("github.com/jfhamlin/muscrat/pkg/aio.WavOut", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.WavOut)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.WithVoices", github_com_jfhamlin_muscrat_pkg_aio.WithVoices)

	// package github.com/jfhamlin/muscrat/pkg/graph
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/graph.Edge", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Edge)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.GeneratorNode", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.GeneratorNode)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.Graph", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Graph)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.Node", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Node)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.NodeID", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.NodeID)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.NodeOption", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.NodeOption)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.SinkChan", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.SinkChan)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.SinkNode", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.SinkNode)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.WithLabel", github_com_jfhamlin_muscrat_pkg_graph.WithLabel)

	// package github.com/jfhamlin/muscrat/pkg/pattern
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/pattern.NewChoose", github_com_jfhamlin_muscrat_pkg_pattern.NewChoose)
	_register("github.com/jfhamlin/muscrat/pkg/pattern.NewSequencer", github_com_jfhamlin_muscrat_pkg_pattern.NewSequencer)

	// package github.com/jfhamlin/freeverb-go
	////////////////////////////////////////
	_register("github.com/jfhamlin/freeverb-go.NewRevModel", github_com_jfhamlin_freeverb_go.NewRevModel)
	_register("github.com/jfhamlin/freeverb-go.RevModel", reflect.TypeOf((*github_com_jfhamlin_freeverb_go.RevModel)(nil)).Elem())

	// package github.com/jfhamlin/muscrat/pkg/slice
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/slice.FindIndexOfRisingEdge", github_com_jfhamlin_muscrat_pkg_slice.FindIndexOfRisingEdge)
}
