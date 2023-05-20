// GENERATED FILE. DO NOT EDIT.
package gljimports

import (
	github_com_jfhamlin_muscrat_pkg_ugen "github.com/jfhamlin/muscrat/pkg/ugen"
	github_com_jfhamlin_muscrat_pkg_wavtabs "github.com/jfhamlin/muscrat/pkg/wavtabs"
	github_com_jfhamlin_muscrat_pkg_stochastic "github.com/jfhamlin/muscrat/pkg/stochastic"
	github_com_jfhamlin_muscrat_pkg_effects "github.com/jfhamlin/muscrat/pkg/effects"
	github_com_jfhamlin_muscrat_pkg_mod "github.com/jfhamlin/muscrat/pkg/mod"
	github_com_jfhamlin_muscrat_pkg_sampler "github.com/jfhamlin/muscrat/pkg/sampler"
	github_com_jfhamlin_muscrat_pkg_aio "github.com/jfhamlin/muscrat/pkg/aio"
	github_com_jfhamlin_muscrat_pkg_graph "github.com/jfhamlin/muscrat/pkg/graph"
	github_com_jfhamlin_freeverb_go "github.com/jfhamlin/freeverb-go"
	"reflect"
)

func RegisterImports(_register func(string, interface{})) {
	// package github.com/jfhamlin/muscrat/pkg/ugen
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewConstant", github_com_jfhamlin_muscrat_pkg_ugen.NewConstant)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewProduct", github_com_jfhamlin_muscrat_pkg_ugen.NewProduct)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewSum", github_com_jfhamlin_muscrat_pkg_ugen.NewSum)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.SampleConfig", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SampleConfig)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.SampleGenerator", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SampleGenerator)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.SampleGeneratorFunc", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SampleGeneratorFunc)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Starter", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Starter)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Stopper", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Stopper)(nil)).Elem())

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
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.NewPinkNoise", github_com_jfhamlin_muscrat_pkg_stochastic.NewPinkNoise)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.Option", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_stochastic.Option)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.PinkNoise", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_stochastic.PinkNoise)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.WithAdd", github_com_jfhamlin_muscrat_pkg_stochastic.WithAdd)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.WithMul", github_com_jfhamlin_muscrat_pkg_stochastic.WithMul)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.WithRand", github_com_jfhamlin_muscrat_pkg_stochastic.WithRand)

	// package github.com/jfhamlin/muscrat/pkg/effects
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewFreeverb", github_com_jfhamlin_muscrat_pkg_effects.NewFreeverb)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewLowpassFilter", github_com_jfhamlin_muscrat_pkg_effects.NewLowpassFilter)

	// package github.com/jfhamlin/muscrat/pkg/mod
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/mod.NewEnvelope", github_com_jfhamlin_muscrat_pkg_mod.NewEnvelope)

	// package github.com/jfhamlin/muscrat/pkg/sampler
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/sampler.LoadSample", github_com_jfhamlin_muscrat_pkg_sampler.LoadSample)
	_register("github.com/jfhamlin/muscrat/pkg/sampler.NewSampler", github_com_jfhamlin_muscrat_pkg_sampler.NewSampler)

	// package github.com/jfhamlin/muscrat/pkg/aio
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/aio.InputDevice", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.InputDevice)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewInputDevice", github_com_jfhamlin_muscrat_pkg_aio.NewInputDevice)

	// package github.com/jfhamlin/muscrat/pkg/graph
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/graph.Edge", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Edge)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.GeneratorNode", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.GeneratorNode)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.Graph", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Graph)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.Node", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Node)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.NodeID", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.NodeID)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.NodeOption", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.NodeOption)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.RunGraph", github_com_jfhamlin_muscrat_pkg_graph.RunGraph)
	_register("github.com/jfhamlin/muscrat/pkg/graph.RunNode", github_com_jfhamlin_muscrat_pkg_graph.RunNode)
	_register("github.com/jfhamlin/muscrat/pkg/graph.SinkChan", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.SinkChan)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.SinkNode", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.SinkNode)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.WithLabel", github_com_jfhamlin_muscrat_pkg_graph.WithLabel)

	// package github.com/jfhamlin/freeverb-go
	////////////////////////////////////////
	_register("github.com/jfhamlin/freeverb-go.NewRevModel", github_com_jfhamlin_freeverb_go.NewRevModel)
	_register("github.com/jfhamlin/freeverb-go.RevModel", reflect.TypeOf((*github_com_jfhamlin_freeverb_go.RevModel)(nil)).Elem())
}
