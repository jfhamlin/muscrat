// GENERATED FILE. DO NOT EDIT.
package gljimports

import (
	"github.com/glojurelang/glojure/pkg/pkgmap"
	github_com_jfhamlin_freeverb_go "github.com/jfhamlin/freeverb-go"
	github_com_jfhamlin_muscrat_pkg_aio "github.com/jfhamlin/muscrat/pkg/aio"
	github_com_jfhamlin_muscrat_pkg_conf "github.com/jfhamlin/muscrat/pkg/conf"
	github_com_jfhamlin_muscrat_pkg_effects "github.com/jfhamlin/muscrat/pkg/effects"
	github_com_jfhamlin_muscrat_pkg_graph "github.com/jfhamlin/muscrat/pkg/graph"
	github_com_jfhamlin_muscrat_pkg_mod "github.com/jfhamlin/muscrat/pkg/mod"
	github_com_jfhamlin_muscrat_pkg_osc "github.com/jfhamlin/muscrat/pkg/osc"
	github_com_jfhamlin_muscrat_pkg_pattern "github.com/jfhamlin/muscrat/pkg/pattern"
	github_com_jfhamlin_muscrat_pkg_sampler "github.com/jfhamlin/muscrat/pkg/sampler"
	github_com_jfhamlin_muscrat_pkg_slice "github.com/jfhamlin/muscrat/pkg/slice"
	github_com_jfhamlin_muscrat_pkg_stochastic "github.com/jfhamlin/muscrat/pkg/stochastic"
	github_com_jfhamlin_muscrat_pkg_ugen "github.com/jfhamlin/muscrat/pkg/ugen"
	github_com_jfhamlin_muscrat_pkg_wavtabs "github.com/jfhamlin/muscrat/pkg/wavtabs"
	"reflect"
)

func init() {
	RegisterImports(pkgmap.Set)
}

func RegisterImports(_register func(string, interface{})) {
	// package github.com/jfhamlin/freeverb-go
	////////////////////////////////////////
	_register("github.com/jfhamlin/freeverb-go.NewRevModel", github_com_jfhamlin_freeverb_go.NewRevModel)
	_register("github.com/jfhamlin/freeverb-go.RevModel", reflect.TypeOf((*github_com_jfhamlin_freeverb_go.RevModel)(nil)).Elem())
	_register("github.com/jfhamlin/freeverb-go.*RevModel", reflect.TypeOf((*github_com_jfhamlin_freeverb_go.RevModel)(nil)))

	// package github.com/jfhamlin/muscrat/pkg/aio
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/aio.InputDevice", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.InputDevice)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.*InputDevice", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.InputDevice)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/aio.Keyboard", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.Keyboard)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.*Keyboard", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.Keyboard)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/aio.KeyboardGate", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.KeyboardGate)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.*KeyboardGate", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.KeyboardGate)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/aio.KeyboardNotes", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.KeyboardNotes)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.*KeyboardNotes", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.KeyboardNotes)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/aio.MIDIControl", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.MIDIControl)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.*MIDIControl", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.MIDIControl)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/aio.MIDIDevice", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.MIDIDevice)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.MIDIDeviceOption", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.MIDIDeviceOption)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.MIDIEnvelope", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.MIDIEnvelope)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.*MIDIEnvelope", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.MIDIEnvelope)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewInputDevice", github_com_jfhamlin_muscrat_pkg_aio.NewInputDevice)
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewMIDIInputDevice", github_com_jfhamlin_muscrat_pkg_aio.NewMIDIInputDevice)
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewQwertyMIDI", github_com_jfhamlin_muscrat_pkg_aio.NewQwertyMIDI)
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewSoftwareKeyboard", github_com_jfhamlin_muscrat_pkg_aio.NewSoftwareKeyboard)
	_register("github.com/jfhamlin/muscrat/pkg/aio.NewWavOut", github_com_jfhamlin_muscrat_pkg_aio.NewWavOut)
	_register("github.com/jfhamlin/muscrat/pkg/aio.QwertyMIDI", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.QwertyMIDI)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.*QwertyMIDI", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.QwertyMIDI)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/aio.SoftwareKeyboard", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.SoftwareKeyboard)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.*SoftwareKeyboard", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.SoftwareKeyboard)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/aio.StdinChan", github_com_jfhamlin_muscrat_pkg_aio.StdinChan)
	_register("github.com/jfhamlin/muscrat/pkg/aio.WavOut", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.WavOut)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/aio.*WavOut", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_aio.WavOut)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/aio.WithChannel", github_com_jfhamlin_muscrat_pkg_aio.WithChannel)
	_register("github.com/jfhamlin/muscrat/pkg/aio.WithController", github_com_jfhamlin_muscrat_pkg_aio.WithController)
	_register("github.com/jfhamlin/muscrat/pkg/aio.WithDefaultValue", github_com_jfhamlin_muscrat_pkg_aio.WithDefaultValue)
	_register("github.com/jfhamlin/muscrat/pkg/aio.WithDeviceID", github_com_jfhamlin_muscrat_pkg_aio.WithDeviceID)
	_register("github.com/jfhamlin/muscrat/pkg/aio.WithDeviceName", github_com_jfhamlin_muscrat_pkg_aio.WithDeviceName)
	_register("github.com/jfhamlin/muscrat/pkg/aio.WithVoices", github_com_jfhamlin_muscrat_pkg_aio.WithVoices)

	// package github.com/jfhamlin/muscrat/pkg/conf
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/conf.BufferSize", github_com_jfhamlin_muscrat_pkg_conf.BufferSize)
	_register("github.com/jfhamlin/muscrat/pkg/conf.SampleFilePaths", github_com_jfhamlin_muscrat_pkg_conf.SampleFilePaths)
	_register("github.com/jfhamlin/muscrat/pkg/conf.SampleRate", github_com_jfhamlin_muscrat_pkg_conf.SampleRate)

	// package github.com/jfhamlin/muscrat/pkg/effects
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/effects.DelayLine", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_effects.DelayLine)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/effects.*DelayLine", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_effects.DelayLine)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewAllPass", github_com_jfhamlin_muscrat_pkg_effects.NewAllPass)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewAmplitude", github_com_jfhamlin_muscrat_pkg_effects.NewAmplitude)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewBPF", github_com_jfhamlin_muscrat_pkg_effects.NewBPF)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewBitcrusher", github_com_jfhamlin_muscrat_pkg_effects.NewBitcrusher)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewClip", github_com_jfhamlin_muscrat_pkg_effects.NewClip)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewDelay", github_com_jfhamlin_muscrat_pkg_effects.NewDelay)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewDelayLine", github_com_jfhamlin_muscrat_pkg_effects.NewDelayLine)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewFreeverb", github_com_jfhamlin_muscrat_pkg_effects.NewFreeverb)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewHiShelf", github_com_jfhamlin_muscrat_pkg_effects.NewHiShelf)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewLimiter", github_com_jfhamlin_muscrat_pkg_effects.NewLimiter)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewLoShelf", github_com_jfhamlin_muscrat_pkg_effects.NewLoShelf)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewLowpassFilter", github_com_jfhamlin_muscrat_pkg_effects.NewLowpassFilter)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewPeakEQ", github_com_jfhamlin_muscrat_pkg_effects.NewPeakEQ)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewPitchShift", github_com_jfhamlin_muscrat_pkg_effects.NewPitchShift)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewRHPF", github_com_jfhamlin_muscrat_pkg_effects.NewRHPF)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewRLPF", github_com_jfhamlin_muscrat_pkg_effects.NewRLPF)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewTapeDelay", github_com_jfhamlin_muscrat_pkg_effects.NewTapeDelay)
	_register("github.com/jfhamlin/muscrat/pkg/effects.NewWaveFolder", github_com_jfhamlin_muscrat_pkg_effects.NewWaveFolder)
	_register("github.com/jfhamlin/muscrat/pkg/effects.WaveFolder", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_effects.WaveFolder)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/effects.*WaveFolder", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_effects.WaveFolder)(nil)))

	// package github.com/jfhamlin/muscrat/pkg/graph
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/graph.AlignGraphs", github_com_jfhamlin_muscrat_pkg_graph.AlignGraphs)
	_register("github.com/jfhamlin/muscrat/pkg/graph.Edge", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Edge)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.*Edge", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Edge)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/graph.Graph", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Graph)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.*Graph", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Graph)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/graph.GraphAlignment", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.GraphAlignment)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.*GraphAlignment", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.GraphAlignment)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/graph.Job", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Job)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.NewQueue", github_com_jfhamlin_muscrat_pkg_graph.NewQueue)
	_register("github.com/jfhamlin/muscrat/pkg/graph.NewRunner", github_com_jfhamlin_muscrat_pkg_graph.NewRunner)
	_register("github.com/jfhamlin/muscrat/pkg/graph.Node", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Node)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.*Node", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Node)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/graph.NodeID", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.NodeID)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.Queue", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Queue)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.*Queue", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Queue)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/graph.QueueItem", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.QueueItem)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.*QueueItem", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.QueueItem)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/graph.Runner", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Runner)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/graph.*Runner", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_graph.Runner)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/graph.SExprToGraph", github_com_jfhamlin_muscrat_pkg_graph.SExprToGraph)

	// package github.com/jfhamlin/muscrat/pkg/mod
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/mod.EnvOption", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_mod.EnvOption)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/mod.NewEnvelope", github_com_jfhamlin_muscrat_pkg_mod.NewEnvelope)
	_register("github.com/jfhamlin/muscrat/pkg/mod.WithCurve", github_com_jfhamlin_muscrat_pkg_mod.WithCurve)
	_register("github.com/jfhamlin/muscrat/pkg/mod.WithReleaseNode", github_com_jfhamlin_muscrat_pkg_mod.WithReleaseNode)

	// package github.com/jfhamlin/muscrat/pkg/osc
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/osc.New", github_com_jfhamlin_muscrat_pkg_osc.New)
	_register("github.com/jfhamlin/muscrat/pkg/osc.NewLFPulse", github_com_jfhamlin_muscrat_pkg_osc.NewLFPulse)
	_register("github.com/jfhamlin/muscrat/pkg/osc.NewLFSaw", github_com_jfhamlin_muscrat_pkg_osc.NewLFSaw)
	_register("github.com/jfhamlin/muscrat/pkg/osc.NewPhasor", github_com_jfhamlin_muscrat_pkg_osc.NewPhasor)
	_register("github.com/jfhamlin/muscrat/pkg/osc.NewPulse", github_com_jfhamlin_muscrat_pkg_osc.NewPulse)
	_register("github.com/jfhamlin/muscrat/pkg/osc.NewSaw", github_com_jfhamlin_muscrat_pkg_osc.NewSaw)
	_register("github.com/jfhamlin/muscrat/pkg/osc.NewSine", github_com_jfhamlin_muscrat_pkg_osc.NewSine)
	_register("github.com/jfhamlin/muscrat/pkg/osc.NewTri", github_com_jfhamlin_muscrat_pkg_osc.NewTri)
	_register("github.com/jfhamlin/muscrat/pkg/osc.Osc", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_osc.Osc)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/osc.*Osc", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_osc.Osc)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/osc.Sampler", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_osc.Sampler)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/osc.SamplerFunc", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_osc.SamplerFunc)(nil)).Elem())

	// package github.com/jfhamlin/muscrat/pkg/pattern
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/pattern.NewChoose", github_com_jfhamlin_muscrat_pkg_pattern.NewChoose)
	_register("github.com/jfhamlin/muscrat/pkg/pattern.NewSequencer", github_com_jfhamlin_muscrat_pkg_pattern.NewSequencer)

	// package github.com/jfhamlin/muscrat/pkg/sampler
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/sampler.LoadSample", github_com_jfhamlin_muscrat_pkg_sampler.LoadSample)
	_register("github.com/jfhamlin/muscrat/pkg/sampler.NewSampler", github_com_jfhamlin_muscrat_pkg_sampler.NewSampler)

	// package github.com/jfhamlin/muscrat/pkg/slice
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/slice.FindIndexOfRisingEdge", github_com_jfhamlin_muscrat_pkg_slice.FindIndexOfRisingEdge)

	// package github.com/jfhamlin/muscrat/pkg/stochastic
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.NewNoise", github_com_jfhamlin_muscrat_pkg_stochastic.NewNoise)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.NewNoiseQuad", github_com_jfhamlin_muscrat_pkg_stochastic.NewNoiseQuad)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.NewPinkNoise", github_com_jfhamlin_muscrat_pkg_stochastic.NewPinkNoise)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.NewRRand", github_com_jfhamlin_muscrat_pkg_stochastic.NewRRand)
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.PinkNoise", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_stochastic.PinkNoise)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/stochastic.*PinkNoise", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_stochastic.PinkNoise)(nil)))

	// package github.com/jfhamlin/muscrat/pkg/ugen
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/ugen.CollectIndexedInputs", github_com_jfhamlin_muscrat_pkg_ugen.CollectIndexedInputs)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.CubInterp", github_com_jfhamlin_muscrat_pkg_ugen.CubInterp)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.DefaultOptions", github_com_jfhamlin_muscrat_pkg_ugen.DefaultOptions)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.GetKnobs", github_com_jfhamlin_muscrat_pkg_ugen.GetKnobs)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Interp", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Interp)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.InterpCubic", github_com_jfhamlin_muscrat_pkg_ugen.InterpCubic)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.InterpLinear", github_com_jfhamlin_muscrat_pkg_ugen.InterpLinear)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.InterpNone", github_com_jfhamlin_muscrat_pkg_ugen.InterpNone)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Knob", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Knob)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.*Knob", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Knob)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/ugen.KnobUpdate", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.KnobUpdate)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.*KnobUpdate", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.KnobUpdate)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/ugen.KnobValueChangeEvent", github_com_jfhamlin_muscrat_pkg_ugen.KnobValueChangeEvent)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.KnobsChangedEvent", github_com_jfhamlin_muscrat_pkg_ugen.KnobsChangedEvent)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.LeakDC", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.LeakDC)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.*LeakDC", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.LeakDC)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/ugen.LinInterp", github_com_jfhamlin_muscrat_pkg_ugen.LinInterp)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewAbs", github_com_jfhamlin_muscrat_pkg_ugen.NewAbs)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewConstant", github_com_jfhamlin_muscrat_pkg_ugen.NewConstant)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewCopySign", github_com_jfhamlin_muscrat_pkg_ugen.NewCopySign)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewExp", github_com_jfhamlin_muscrat_pkg_ugen.NewExp)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewFMA", github_com_jfhamlin_muscrat_pkg_ugen.NewFMA)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewFMAStatic", github_com_jfhamlin_muscrat_pkg_ugen.NewFMAStatic)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewFreqRatio", github_com_jfhamlin_muscrat_pkg_ugen.NewFreqRatio)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewHydra", github_com_jfhamlin_muscrat_pkg_ugen.NewHydra)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewImpulse", github_com_jfhamlin_muscrat_pkg_ugen.NewImpulse)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewKnob", github_com_jfhamlin_muscrat_pkg_ugen.NewKnob)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewLatch", github_com_jfhamlin_muscrat_pkg_ugen.NewLatch)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewLinExp", github_com_jfhamlin_muscrat_pkg_ugen.NewLinExp)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewLog2", github_com_jfhamlin_muscrat_pkg_ugen.NewLog2)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewMIDIFreq", github_com_jfhamlin_muscrat_pkg_ugen.NewMIDIFreq)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewMax", github_com_jfhamlin_muscrat_pkg_ugen.NewMax)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewMin", github_com_jfhamlin_muscrat_pkg_ugen.NewMin)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewMovingAverage", github_com_jfhamlin_muscrat_pkg_ugen.NewMovingAverage)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewPow", github_com_jfhamlin_muscrat_pkg_ugen.NewPow)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewProduct", github_com_jfhamlin_muscrat_pkg_ugen.NewProduct)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewPulseDiv", github_com_jfhamlin_muscrat_pkg_ugen.NewPulseDiv)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewQuotient", github_com_jfhamlin_muscrat_pkg_ugen.NewQuotient)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewScope", github_com_jfhamlin_muscrat_pkg_ugen.NewScope)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewSine", github_com_jfhamlin_muscrat_pkg_ugen.NewSine)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewSum", github_com_jfhamlin_muscrat_pkg_ugen.NewSum)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NewTanh", github_com_jfhamlin_muscrat_pkg_ugen.NewTanh)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.NextPowerOf2", github_com_jfhamlin_muscrat_pkg_ugen.NextPowerOf2)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Option", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Option)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Options", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Options)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.*Options", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Options)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/ugen.SampleConfig", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SampleConfig)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.*SampleConfig", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SampleConfig)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Scope", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Scope)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.*Scope", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Scope)(nil)))
	_register("github.com/jfhamlin/muscrat/pkg/ugen.SimpleUGenFunc", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.SimpleUGenFunc)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Starter", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Starter)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Stopper", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.Stopper)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.UGen", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.UGen)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.UGenFunc", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_ugen.UGenFunc)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/ugen.WithDefaultDutyCycle", github_com_jfhamlin_muscrat_pkg_ugen.WithDefaultDutyCycle)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.WithInterp", github_com_jfhamlin_muscrat_pkg_ugen.WithInterp)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.WithRand", github_com_jfhamlin_muscrat_pkg_ugen.WithRand)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.WithSeed", github_com_jfhamlin_muscrat_pkg_ugen.WithSeed)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.ZapGremlins", github_com_jfhamlin_muscrat_pkg_ugen.ZapGremlins)
	_register("github.com/jfhamlin/muscrat/pkg/ugen.Zeros", github_com_jfhamlin_muscrat_pkg_ugen.Zeros)

	// package github.com/jfhamlin/muscrat/pkg/wavtabs
	////////////////////////////////////////
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.DefaultResolution", github_com_jfhamlin_muscrat_pkg_wavtabs.DefaultResolution)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.New", github_com_jfhamlin_muscrat_pkg_wavtabs.New)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.NewWithWrap", github_com_jfhamlin_muscrat_pkg_wavtabs.NewWithWrap)
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.Table", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_wavtabs.Table)(nil)).Elem())
	_register("github.com/jfhamlin/muscrat/pkg/wavtabs.*Table", reflect.TypeOf((*github_com_jfhamlin_muscrat_pkg_wavtabs.Table)(nil)))
}
