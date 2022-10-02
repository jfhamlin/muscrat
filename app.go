package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/notes"
	"github.com/jfhamlin/muscrat/internal/pkg/wavtabs"

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

	outputChannel      chan []float64
	graphOutputChannel <-chan []float64

	synthFileName string
	sampleRate    int

	cancelGraph func()
	cancelSink  func()
	graph       *graph.Graph

	mtx sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		outputChannel: make(chan []float64, 4), // buffer four packets of samples
		synthFileName: "synth.mrat",
		gain:          0.25,
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

func harmonizer(gengen func(hz float64) generator.SampleGenerator, rootHz float64, numGens int) generator.SampleGenerator {
	gens := make([]generator.SampleGenerator, numGens)
	weights := make([]float64, numGens)

	for i := 0; i < numGens; i++ {
		gens[i] = gengen(rootHz * float64(i+1))
		weights[i] = 1.0 / float64(i+1)
	}
	return NewSampleGeneratorSet(gens, weights)
}

func sineHarmonizer(rootHz float64) generator.SampleGenerator {
	const numGens = 1
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
	return wavtabGenerator(wavtabs.Sin(1024), hz)
}

func wavtabGenerator(wavtab wavtabs.Table, hz float64) generator.SampleGenerator {
	phase := 0.0
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := 0; i < n; i++ {
			res[i] = wavtab.Lerp(phase)
			phase += float64(hz) / float64(cfg.SampleRateHz)
			if phase > 1 {
				phase -= 1
			}
		}
		return res
	})
}

var Noise = generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
	res := make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = 2*rand.Float64() - 1
	}
	return res
})

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
			fmt.Println("XXX clipping high")
		}
		if s < 0 {
			fmt.Println("XXX clipping low")
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
	a.sampleRate = cfg.SampleRate

	// set up the audio output
	sink, err := sinks.NewSDLSink(cfg)
	if err != nil {
		panic(err)
	}
	sink.Start(a.getSamples)

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
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op == fsnotify.Write {
					a.updateSignalGraphFromScriptFile(a.synthFileName)
				}
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
	g, sinkChannels, err := scriptToGraph(string(synthFile))
	if err != nil {
		fmt.Println("error parsing script:", err)
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
		const fadeTime = 10 * time.Millisecond
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
		fmt.Println("finished fade")

		// stop the old graph and wait for it to finish
		a.cancelGraph()
	}

	go func() {
		defer close(sinkStopChan)
		for {
			select {
			case <-sinkCtx.Done():
				fmt.Println("sink context done")
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

type command struct {
	oscName string
	oscArgs map[string]interface{}
	freq    float64
	amp     float64
}

func parseCommand(cmd string) (command, error) {
	// parse the command. It should be of the form:
	// <sin|saw|sqr|noise> [(OSC_ARGS)] <freq|midi note name> [<amp>]
	// where OSC_ARGS is a comma-separated list of key=value pairs

	// preprocess the command to add whitespace before and after allowed
	// non-alphanumeric characters
	re := regexp.MustCompile(`[(,)=]`)
	cmd = re.ReplaceAllString(cmd, " $0 ")

	var res command

	fields := strings.Fields(cmd)
	if len(fields) < 2 {
		return res, fmt.Errorf("invalid command: %s", cmd)
	}

	res.oscName = fields[0]
	fields = fields[1:]

	res.oscArgs = make(map[string]interface{})
	if fields[0] == "(" {
		fields = fields[1:]
		for fields[0] != ")" && len(fields) > 0 {
			if len(fields) < 3 {
				return res, fmt.Errorf("invalid command: %s", cmd)
			}
			key := fields[0]
			if fields[1] != "=" {
				return res, fmt.Errorf("invalid command: %s", cmd)
			}
			value := fields[2]
			res.oscArgs[key] = value
			fields = fields[3:]

			if fields[0] == "," {
				fields = fields[1:]
			}
		}
		if fields[0] != ")" {
			return res, fmt.Errorf("invalid command, expected ')': %s", cmd)
		}
		fields = fields[1:]
	}

	if len(fields) < 1 {
		return res, fmt.Errorf("invalid command, missing frequency: %s", cmd)
	}

	freq, err := parseFreq(fields[0])
	if err != nil {
		return res, fmt.Errorf("invalid command, invalid frequency: %s", cmd)
	}
	res.freq = freq
	fields = fields[1:]

	if len(fields) == 0 {
		res.amp = 1
		return res, nil
	}
	amp, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return res, fmt.Errorf("invalid command, invalid amplitude: %s", cmd)
	}
	res.amp = amp

	if len(fields) > 1 {
		return res, fmt.Errorf("invalid command, too many fields: %s", cmd)
	}

	return res, nil
}

func parseFreq(note string) (float64, error) {
	if note[0] >= '0' && note[0] <= '9' {
		return strconv.ParseFloat(note, 64)
	} else {
		return noteNameToFreq(note)
	}
}

func scriptToGraph(synthDesc string) (g *graph.Graph, sinks []<-chan []float64, err error) {
	// split synth file into lines
	lines := strings.Split(synthDesc, "\n")
	commands := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}
		commands = append(commands, line)
	}

	g = &graph.Graph{}
	sinkID, graphOutputChannel := g.AddSinkNode()
	var oscNodeIDs []graph.NodeID

	for _, cmd := range commands {
		command, cmdErr := parseCommand(cmd)
		if err != nil {
			return nil, nil, cmdErr
		}
		// parse the command. It should be of the form:
		// <sin|saw|sqr|noise>[(<key>=<value>[, ...])] <freq|midi note name> [<amp>]
		/*
			parts := strings.Split(cmd, " ")
			if len(parts) < 2 {
				err = fmt.Errorf("invalid command: %s", cmd)
				return
			}
			if len(parts) > 3 {
				err = fmt.Errorf("too many arguments: %s", cmd)
				return
			}
			if len(parts) == 2 {
				parts = append(parts, "1")
			}
			osc, note, ampStr := parts[0], parts[1], parts[2]

			// parse the note
			var freq float64
			if note[0] >= '0' && note[0] <= '9' {
				// it's a frequency
				freq, err = strconv.ParseFloat(note, 64)
				if err != nil {
					err = fmt.Errorf("invalid frequency: %s", note)
					return
				}
			} else {
				// it's a note name
				freq, err = noteNameToFreq(note)
				if err != nil {
					err = fmt.Errorf("invalid note name: %s", note)
					return
				}
			}
			var amp float64
			amp, err = strconv.ParseFloat(ampStr, 64)
			if err != nil {
				err = fmt.Errorf("invalid amplitude: %s", ampStr)
				return
			}
		*/
		var genFunc func(freq float64) generator.SampleGenerator

		switch command.oscName {
		case "sin":
			genFunc = func(freq float64) generator.SampleGenerator {
				return wavtabGenerator(wavtabs.Sin(1024), freq)
			}
		case "saw":
			genFunc = func(freq float64) generator.SampleGenerator {
				return wavtabGenerator(wavtabs.Saw(1024), freq)
			}
		case "sqr":
			defaultDutyCycle := "0.5"
			dutyCycle, ok := command.oscArgs["dc"]
			if !ok {
				dutyCycle = defaultDutyCycle
			}
			dc, err := strconv.ParseFloat(dutyCycle.(string), 64)
			if err != nil {
				err = fmt.Errorf("invalid duty cycle: %s", dutyCycle)
				return nil, nil, err
			}
			genFunc = func(freq float64) generator.SampleGenerator {
				return wavtabGenerator(wavtabs.Square(1024, dc), freq)
			}
		case "noise":
			genFunc = func(freq float64) generator.SampleGenerator {
				return Noise
			}
		default:
			err = fmt.Errorf("invalid oscillator: %s", command.oscName)
			return
		}
		genID := g.AddGeneratorNode(genFunc(command.freq))

		amp := command.amp
		if amp == 1 {
			oscNodeIDs = append(oscNodeIDs, genID)
			continue
		}
		ampID := g.AddGeneratorNode(generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
			res := make([]float64, n)
			for i := range res {
				res[i] = amp * cfg.InputSamples[0][i]
			}
			return res
		}))
		g.AddEdge(genID, ampID)

		oscNodeIDs = append(oscNodeIDs, ampID)

		// //gen := sineHarmonizer(notes.GetNote(note).Frequency)
		// gen := harmonizer(func(hz float64) generator.SampleGenerator {
		// 	return wavtabGenerator(wavtabs.Square(1024, 0.8), hz)
		// }, notes.GetNote(note).Frequency, 3)
		// //gen := wavtabGenerator(wavtabs.Square(1024, 0.8), notes.GetNote(note).Frequency)
		// id := g.AddGeneratorNode(gen)
		// oscNodeIDs = append(oscNodeIDs, id)
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
	for _, id := range oscNodeIDs {
		g.AddEdge(id, mixerNodeID)
	}
	fmt.Println(len(g.Nodes), "nodes")
	for _, e := range g.Edges {
		fmt.Printf("%v -> %v\n", e.From, e.To)
	}

	return g, []<-chan []float64{graphOutputChannel}, nil
}

func noteNameToFreq(name string) (float64, error) {
	note := notes.GetNote(name)
	if note == nil {
		return 0, fmt.Errorf("invalid note name: %s", name)
	}
	return note.Frequency, nil
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
	a.waveformCallback = cb
}
