package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/cmplx"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/mjibson/go-dsp/fft"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"golang.org/x/term"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang"
	"github.com/jfhamlin/muscrat/internal/pkg/notes"
	"github.com/jfhamlin/muscrat/internal/pkg/plot"

	"github.com/bspaans/bleep/audio"
	"github.com/bspaans/bleep/sinks"

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

	showSpectrum     bool
	showSpectrumHist bool

	showOscilloscope         bool
	oscilloscopeWindow       float64
	oscilloscopeUpdateFreqHz float64

	waveformCallback WaveformCallback

	outputChannel      chan []float64
	graphOutputChannel <-chan []float64

	synthFileName string
	sampleRate    int

	cancelGraph       func()
	cancelSink        func()
	graph             *graph.Graph
	lastSynthFileHash [32]byte

	mtx sync.RWMutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		outputChannel: make(chan []float64, 4), // buffer four packets of samples
		synthFileName: "synth.mrat",
		gain:          0.25,
		// showSpectrum:             true,
		showSpectrumHist: true,
		// showOscilloscope:         true,
		oscilloscopeWindow:       1.0 / 440,
		oscilloscopeUpdateFreqHz: 1,
	}
}

func transformSampleBuffer(cfg *audio.AudioConfig, buf []float64) []int {
	maxValue := float64(int(1) << cfg.BitDepth)

	var out []int
	if cfg.Stereo {
		out = make([]int, 2*len(buf))
	} else {
		out = make([]int, len(buf))
	}

	for i, sample := range buf {
		s := (sample + 1) * (maxValue / 2)
		if s > maxValue {
			//fmt.Printf("XXX clipping high (max=%v): %v (%v)\n", maxValue, s, sample)
		}
		if s < 0 {
			//fmt.Printf("XXX clipping low (min=%v): %v (%v)\n", 0, s, sample)
		}
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
	grayscaleASCII   = []byte(" .:-=+*#%@")
	grayscaleASCII70 = []byte(" .'`^\",:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$")

	fftOn    = true
	timingOn = false

	fftChan = make(chan []float64, 2)
)

const (
	// TODO: don't hardcode this.
	sampleRate = 44100
)

func renderSpectrumLine(x []complex128, width int) string {
	abs := make([]float64, width)
	// group pairs of bins. ignore the top half of the spectrum.
	for i := 0; i < len(x)/2; i++ {
		// map to [0, 255] in abs with a logaritmic scale so that more
		// resolution is given to the lower frequencies.
		t := float64(len(abs)-1) * math.Log10(float64(i+1)) / math.Log10(float64(len(x)/2))
		floorT := float64(int(t))
		t -= floorT
		val := cmplx.Abs(x[i])
		abs[int(floorT)] += (1 - t) * val
		if floorT < float64(len(abs)-1) {
			abs[int(floorT)+1] += t * val
		}
	}

	for i := 0; i < len(abs); i++ {
		abs[i] = math.Log10(abs[i] + 1)
	}

	maxVal := 0.0
	for i := 0; i < len(abs); i++ {
		maxVal = math.Max(maxVal, abs[i])
	}
	const lines = 10

	builder := strings.Builder{}
	builder.WriteRune('|')
	for i := 0; i < len(abs); i++ {
		val := abs[i]
		if val < 0.1 {
			val = 0
		}
		builder.WriteByte(grayscaleASCII70[int((val/maxVal)*float64(len(grayscaleASCII70)-1))])
	}
	builder.WriteRune('|')
	return builder.String()
}

// renderSpectrumHist renders an ASCII histogram of the spectrum with
// a logarithmic scale. The bottom line shows frequency labels.
func renderSpectrumHist(x []complex128, width, height int) string {
	return plot.DFTHistogramString(x, sampleRate, width, height, plot.WithLogDomain(), plot.WithLogRange(), plot.WithMinFreq(20), plot.WithMaxFreq(20000))
}

func (a *App) renderOscilloscope(samps []float64, width, height int) string {
	return plot.LineChartString(samps, width, height)
}

func (a *App) spectrumWorker() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Println("not a terminal")
		return
	}

	lastOscilloscopeRender := time.Now()
	var lastOscilloscopeFrame string
	for {
		func() {
			oscilloscopeRenderInterval := time.Duration(1 / a.oscilloscopeUpdateFreqHz * float64(time.Second))

			width, height, err := term.GetSize(int(os.Stdout.Fd()))
			if err != nil || (!a.showOscilloscope && !a.showSpectrum) {
				return
			}

			samps := <-fftChan
			samps = append(samps, (<-fftChan)...)

			builder := strings.Builder{}

			builder.WriteRune('\n')
			if a.showOscilloscope {
				oHeight := height
				if a.showSpectrum {
					oHeight = height / 2
					height -= oHeight
				}

				if time.Since(lastOscilloscopeRender) >= oscilloscopeRenderInterval {
					lastOscilloscopeRender = time.Now()
					oscNumSamples := int(a.oscilloscopeWindow * 44100)
					if oscNumSamples == 0 {
						fmt.Println("oscilloscope window too small:", a.oscilloscopeWindow)
						return
					}
					for len(samps) < oscNumSamples {
						samps = append(samps, (<-fftChan)...)
					}

					lastOscilloscopeFrame = a.renderOscilloscope(samps[:oscNumSamples], width, oHeight)
				}
				builder.WriteString(lastOscilloscopeFrame)
			}

			if !a.showSpectrum {
				fmt.Print(builder.String())
				return
			}

			if a.showOscilloscope {
				builder.WriteRune('\n')
			}
			// truncate the length of samps to the nearest power of 2.
			samps = samps[:1<<uint(math.Log2(float64(len(samps))))]

			// apply a Bartlett window to the samples to reduce the spectral
			// leakage.
			for i := 0; i < len(samps); i++ {
				if i < (len(samps)-1)/2 {
					samps[i] *= 2 * float64(i) / float64(len(samps)-1)
				} else {
					samps[i] *= 2 - 2*float64(i)/float64(len(samps)-1)
				}
			}
			x := fft.FFTReal(samps)
			if a.showSpectrumHist {
				builder.WriteString(renderSpectrumHist(x, width, height))
			} else {
				builder.WriteString(renderSpectrumLine(x, width-2))
			}
			fmt.Print(builder.String())
		}()
	}
}

func (a *App) getSamples(cfg *audio.AudioConfig, n int) []int {
	if timingOn {
		start := time.Now()
		defer func() {
			dur := time.Since(start)
			budget := time.Second * time.Duration(n) / time.Duration(cfg.SampleRate)
			fmt.Printf("[getSamples] duration=%v budget overage=%v\n", dur, dur-budget)
		}()
	}

	var samples []float64
	select {
	case samples = <-a.outputChannel:
		// return silence if we can't get samples fast enough.
	case <-time.After(time.Duration(n) * time.Second / time.Duration(cfg.SampleRate)):
		samples = make([]float64, n)
	}

	select {
	case fftChan <- samples:
	default:
	}

	go runtime.EventsEmit(a.ctx, "samples", samples)

	samples = scaleSamples(samples, a.gain)
	return transformSampleBuffer(cfg, samples)
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	cfg := audio.NewAudioConfig()
	a.sampleRate = cfg.SampleRate

	// set up the audio output
	sink, err := sinks.NewSDLSink(cfg)
	if err != nil {
		panic(err)
	}
	sink.Start(a.getSamples)

	go a.spectrumWorker()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	if err := watcher.Add(a.synthFileName); err != nil {
		panic(err)
	}

	go func() {
		defer watcher.Close()

		for {
			select {
			case _, ok := <-watcher.Events:
				if !ok {
					return
				}
				a.updateSignalGraphFromScriptFile(a.synthFileName)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("error:", err)
				// TODO: handle program termination
			}
		}
	}()

	a.updateSignalGraphFromScriptFile(a.synthFileName)
}

func (a *App) updateSignalGraphFromScriptFile(filename string) {
	// read the synth file
	synthFile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	// hash the file contents
	hash := sha256.Sum256(synthFile)
	// if the hash is the same as the last time we loaded the file, don't do anything
	if bytes.Equal(hash[:], a.lastSynthFileHash[:]) {
		return
	}
	a.lastSynthFileHash = hash

	program, err := mratlang.Parse(strings.NewReader(string(synthFile)), mratlang.WithFilename(filename))
	if err != nil {
		fmt.Println("error parsing script:", err)
		return
	}
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	g, sinkChannels, err := func() (g *graph.Graph, sc []graph.SinkChan, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v\n%s", r, debug.Stack())
			}
		}()
		return program.Eval(mratlang.WithLoadPath([]string{pwd}))
	}()
	if err != nil {
		fmt.Println("error generating graph:", err)
		return
	}
	if len(sinkChannels) == 0 {
		fmt.Println("no sink channels found")
		return
	}
	graphOutputChannel := sinkChannels[0]

	// if we already have a graph, stop it, then start the new one.
	a.mtx.Lock()
	defer a.mtx.Unlock()

	graphStopChan := make(chan struct{})
	sinkStopChan := make(chan struct{})
	graphCtx, cancelGraph := context.WithCancel(context.Background())
	sinkCtx, cancelSink := context.WithCancel(context.Background())

	go func() {
		defer close(graphStopChan)
		graph.RunGraph(graphCtx, g, generator.SampleConfig{SampleRateHz: a.sampleRate})
	}()

	if a.graph != nil {
		// stop the old sink goroutine
		a.cancelSink()

		// start the fade between the old and new graphs
		const fadeTime = 100 * time.Millisecond
		startTime := time.Now()
		fmt.Println("starting fade")
		for time.Since(startTime) < fadeTime {
			sinceStart := time.Since(startTime)

			samplesOld := <-a.graphOutputChannel
			samplesNew := <-graphOutputChannel

			samplesMixed := make([]float64, len(samplesOld))
			for i := range samplesOld {
				samplesMixed[i] = samplesOld[i]*(1-sinceStart.Seconds()/fadeTime.Seconds()) + samplesNew[i]*(sinceStart.Seconds()/fadeTime.Seconds())
			}
			a.outputChannel <- samplesMixed
		}

		// stop the old graph and wait for it to finish
		a.cancelGraph()
	}

	go func() {
		defer close(sinkStopChan)
		for {
			select {
			case <-sinkCtx.Done():
				return
			case samples, ok := <-a.graphOutputChannel:
				if !ok {
					return
				}
				a.outputChannel <- samples
			}
		}
	}()

	// update state
	a.cancelGraph = func() {
		cancelGraph()
		<-graphStopChan
	}
	a.cancelSink = func() {
		cancelSink()
		<-sinkStopChan
	}
	a.graph = g
	a.graphOutputChannel = graphOutputChannel
}

func (a *App) SetGain(gain float64) {
	a.gain = math.Max(0, math.Min(gain, 1))
}

func (a *App) GetNotes() []string {
	return notes.Names()
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
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.waveformCallback = cb
}

func (a *App) GraphDot() string {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	if a.graph == nil {
		return ""
	}
	return a.graph.Dot()
}

func (a *App) GraphJSON() string {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	if a.graph == nil {
		return ""
	}

	buf, err := json.Marshal(a.graph)
	if err != nil {
		panic(err)
	}
	return string(buf)
}

func (a *App) SetShowSpectrum(showSpectrum bool) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.showSpectrum = showSpectrum
}

func (a *App) SetShowSpectrumHist(showSpectrumHist bool) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.showSpectrumHist = showSpectrumHist
}

func (a *App) SetShowOscilloscope(showOscilloscope bool) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.showOscilloscope = showOscilloscope
}

func (a *App) SetOscilloscopeWindow(window float64) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.oscilloscopeWindow = window
}

func (a *App) SetOscilloscopeFreq(freq float64) {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.oscilloscopeUpdateFreqHz = freq
}
