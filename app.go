package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/notes"

	"github.com/bspaans/bleep/audio"
	"github.com/bspaans/bleep/sinks"

	"github.com/mjibson/go-dsp/spectral"

	"net/http"
	_ "net/http/pprof"
)

func init() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

// App struct
type App struct {
	ctx context.Context

	gain float64

	generator     generator.SampleGenerator
	nextGenerator generator.SampleGenerator
	fade          float64

	fftBuffer []float64

	waveformCallback WaveformCallback

	outputChannel chan []float64

	note int

	mtx sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		outputChannel: make(chan []float64, 1),
	}
}

// SampleGeneratorSet is a sample generator that sums the outputs of
// multiple sample generators.
type SampleGeneratorSet struct {
	Generators []generator.SampleGenerator
	Weights    []float64
}

func (s *SampleGeneratorSet) GenerateSamples(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
	res := make([]float64, n)
	for i, g := range s.Generators {
		samples := g.GenerateSamples(ctx, cfg, n)
		for j := 0; j < n; j++ {
			res[j] += s.Weights[i] * samples[j]
		}
	}
	return res
}

func NewSampleGeneratorSet(generators []generator.SampleGenerator, weights []float64) *SampleGeneratorSet {
	if len(weights) > len(generators) {
		weights = weights[:len(generators)]
	}
	for len(weights) < len(generators) {
		weights = append(weights, 1)
	}
	var sum float64
	for i, w := range weights {
		if w < 0 {
			w *= -1
			weights[i] = w
		}
		sum += w
	}
	for i := range weights {
		weights[i] /= sum
	}
	return &SampleGeneratorSet{
		Generators: generators,
		Weights:    weights,
	}
}

// SampleScaler is a sample generator that scales the output of another.
type SampleScaler struct {
	Generator generator.SampleGenerator
	Gain      float64
}

func (s *SampleScaler) GenerateSamples(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
	samples := s.Generator.GenerateSamples(ctx, cfg, n)
	for i := 0; i < n; i++ {
		samples[i] *= s.Gain
	}
	return samples
}

func NewSampleScaler(g generator.SampleGenerator, gain float64) *SampleScaler {
	return &SampleScaler{
		Generator: g,
		Gain:      gain,
	}
}

func sineHarmonizer(rootHz float64) generator.SampleGenerator {
	const numGens = 2
	sines := make([]generator.SampleGenerator, numGens)
	weights := make([]float64, numGens)

	sines[0] = sineGenerator(rootHz)
	weights[0] = 1
	for i := 1; i < numGens; i++ {
		sines[i] = NewSampleScaler(sineGenerator(rootHz*float64(i+1)), 1.0/float64(i+1))
		weights[i] = 1.0 / float64(i+1)
	}
	return NewSampleGeneratorSet(sines, weights)
}

func sineGenerator(hz float64) generator.SampleGenerator {
	phase := 0.0
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := 0; i < n; i++ {
			res[i] = math.Sin(phase)
			phase += 2 * math.Pi * float64(hz) / float64(cfg.SampleRateHz)
		}
		return res
	})
}

func noiseGenerator(cfg generator.SampleConfig, n int) []float64 {
	res := make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = rand.Float64()
	}
	return res
}

func transformSampleBuffer(cfg *audio.AudioConfig, buf []float64) []int {
	maxValue := math.Pow(2, float64(cfg.BitDepth))

	var out []int
	if cfg.Stereo {
		out = make([]int, 2*len(buf))
	} else {
		out = make([]int, len(buf))
	}

	for i, sample := range buf {
		s := (sample + 1) * (maxValue / 2)
		s = math.Max(0, math.Ceil(s))
		sout := int(math.Min(s, maxValue-1))

		if cfg.Stereo {
			out[2*i] = sout
			out[2*i+1] = sout
		} else {
			out[i] = sout
		}
	}

	return out
}

func scaleSamples(buf []float64, gain float64) []float64 {
	res := make([]float64, len(buf))
	for i := 0; i < len(buf); i++ {
		res[i] = buf[i] * gain
	}
	return res
}

var (
	lastSpectrumCheck time.Time
)

func (a *App) getSamples(cfg *audio.AudioConfig, n int) []int {
	samples := <-a.outputChannel
	samples = scaleSamples(samples, a.gain)
	return transformSampleBuffer(cfg, samples)
}

func (a *App) getSamplesOld(cfg *audio.AudioConfig, n int) []int {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	var samps []float64
	if a.generator != nil {
		samps = a.generator.GenerateSamples(context.Background(), generator.SampleConfig{SampleRateHz: cfg.SampleRate}, n)
	} else {
		samps = make([]float64, n)
	}

	if a.nextGenerator != nil {
		const fadeDuration = 0.2
		nextSamps := a.nextGenerator.GenerateSamples(context.Background(), generator.SampleConfig{SampleRateHz: cfg.SampleRate}, n)
		// fade between the two generators until fade is 1
		for i := 0; i < n; i++ {
			samps[i] = samps[i]*(1-a.fade) + nextSamps[i]*a.fade
			a.fade += 1.0 / (fadeDuration * float64(cfg.SampleRate))
			if a.fade >= 1 {
				a.fade = 1
			}
		}
		if a.fade >= 1 {
			a.generator = a.nextGenerator
			a.nextGenerator = nil
			a.fade = 0
		}
	}

	samps = scaleSamples(samps, a.gain)

	a.fftBuffer = append(a.fftBuffer, samps...)
	if len(a.fftBuffer) > 1024*4 {
		a.fftBuffer = a.fftBuffer[len(a.fftBuffer)-1024*4:]
	}
	if time.Since(lastSpectrumCheck) > 2*time.Second {
		pow, freqs := spectral.Pwelch(samps, float64(cfg.SampleRate), &spectral.PwelchOptions{})
		res := make([]struct{ Freq, Power float64 }, len(pow))
		var powerSum float64
		for i := range pow {
			res[i] = struct{ Freq, Power float64 }{freqs[i], pow[i]}
			powerSum += pow[i]
		}
		sort.Slice(res, func(i, j int) bool {
			return res[i].Power > res[j].Power
		})

		lastSpectrumCheck = time.Now()
		a.fftBuffer = a.fftBuffer[:0]
	}

	return transformSampleBuffer(cfg, samps)
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg := audio.NewAudioConfig()

	sink, err := sinks.NewSDLSink(cfg)
	if err != nil {
		panic(err)
	}
	sink.Start(a.getSamples)

	g := &graph.Graph{}
	sinkID, outputChannel := g.AddSinkNode()
	var sineNodeIDs []graph.NodeID
	for _, note := range []string{"Gs2", "Db3", "Gs3", "As3", "Ab3", "Gs4"} {
		id := g.AddGeneratorNode(sineHarmonizer(notes.GetNote(note).Frequency))
		sineNodeIDs = append(sineNodeIDs, id)
	}
	mixerNodeID := g.AddGeneratorNode(generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for _, samples := range cfg.InputSamples {
			for i := 0; i < n; i++ {
				res[i] += samples[i] / float64(len(cfg.InputSamples))
			}
		}
		return res
	}))
	g.AddEdge(mixerNodeID, sinkID)
	for _, id := range sineNodeIDs {
		g.AddEdge(id, mixerNodeID)
	}
	fmt.Println(len(g.Nodes), "nodes")
	for _, e := range g.Edges {
		fmt.Printf("%v -> %v\n", e.From, e.To)
	}
	go func() {
		for samples := range outputChannel {
			a.outputChannel <- samples
		}
	}()
	go graph.RunGraph(ctx, g, generator.SampleConfig{SampleRateHz: cfg.SampleRate})
}

func (a *App) SetGain(gain float64) {
	a.gain = math.Max(0, math.Min(gain, 1))
}

func (a *App) GetNotes() []string {
	return notes.Names()
}

func (a *App) SetChord(noteNames []string, noteWeights []float64) {
	gens := make([]generator.SampleGenerator, len(noteNames))
	for i, noteName := range noteNames {
		note := notes.GetNote(noteName)
		if note == nil {
			fmt.Println("unknown note", noteName)
			return
		}
		gens[i] = sineHarmonizer(note.Frequency)
	}

	generatorSet := NewSampleGeneratorSet(gens, noteWeights)

	a.mtx.Lock()
	defer a.mtx.Unlock()

	a.nextGenerator = generatorSet
}

type FreqInfo struct {
	Freq  float64
	Power float64
}

type WaveformInfo struct {
	SampleRateHz   float64
	Samples        []float64
	FrequencyPower []FreqInfo
}

type WaveformCallback func(samples []float64, sampleFreq float64)

func (a *App) RegisterWaveformCallback(cb WaveformCallback) {
	a.waveformCallback = cb
}
