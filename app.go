package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/fsnotify/fsnotify"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang"
	"github.com/jfhamlin/muscrat/internal/pkg/notes"
	"github.com/jfhamlin/muscrat/internal/pkg/wavtabs"

	"github.com/jfhamlin/muscrat/pkg/freeverb"

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

	cancelGraph       func()
	cancelSink        func()
	graph             *graph.Graph
	lastSynthFileHash [32]byte

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

// func sineHarmonizer(rootHz float64) generator.SampleGenerator {
// 	const numGens = 1
// 	sines := make([]generator.SampleGenerator, numGens)
// 	weights := make([]float64, numGens)

// 	sines[0] = sineGenerator()
// 	weights[0] = 1
// 	for i := 1; i < numGens; i++ {
// 		sines[i] = NewSampleScaler(sineGenerator(rootHz*float64(i+1)), 1.0/float64(i+1))
// 		weights[i] = 1.0 / float64(i+1)
// 	}
// 	return NewSampleGeneratorSet(sines, weights)
// }

func sineGenerator() generator.SampleGenerator {
	return wavtabGenerator(wavtabs.Sin(1024))
}

func wavtabGenerator(wavtab wavtabs.Table) generator.SampleGenerator {
	phase := 0.0
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		ws := cfg.InputSamples["w"]
		res := make([]float64, n)
		// default frequency; use the last value if we run out of
		// input samples
		w := 0.0
		for i := 0; i < n; i++ {
			if i < len(ws) {
				w = ws[i]
			}
			res[i] = wavtab.Lerp(phase)
			phase += w / float64(cfg.SampleRateHz)
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

func NewADSRGenerator() generator.SampleGenerator {
	type adsrState int
	const (
		stateOff adsrState = iota
		stateAttack
		stateDecay
		stateSustain
		stateRelease
	)
	state := stateOff
	stateTime := 0.0
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		samples := make([]float64, n)
		trigger := cfg.InputSamples["trigger"]
		attacks := cfg.InputSamples["a"]
		decays := cfg.InputSamples["d"]
		sustains := cfg.InputSamples["s"]
		releases := cfg.InputSamples["r"]

		for i := 0; i < n; i++ {
			attack := attacks[i]
			decay := decays[i]
			sustain := sustains[i]
			release := releases[i]

			// first, handle state transitions
			nextState := state
			switch state {
			case stateOff:
				if trigger[i] > 0 {
					nextState = stateAttack
				}
			case stateAttack:
				if stateTime >= attack {
					nextState = stateDecay
				}
			case stateDecay:
				if stateTime >= decay {
					if trigger[i] > 0 {
						nextState = stateSustain
					} else {
						nextState = stateRelease
					}
				}
			case stateSustain:
				if trigger[i] <= 0 {
					nextState = stateRelease
				}
			case stateRelease:
				switch {
				case trigger[i] > 0:
					nextState = stateAttack
				case stateTime >= release:
					nextState = stateOff
				}
			default:
				panic("unreachable adsr state")
			}
			if nextState != state {
				if state == stateRelease && nextState == stateAttack {
					// if we're transitioning from release to attack, we need to
					// set the stateTime to match the output level to start from
					// the current amplitude.
					//
					// a little algebra, setting the output levels to be equal:
					// out_attack = stateTime_a / attack
					// out_release = sustain * (1 - stateTime_r/release)
					// out_attack = out_release
					// stateTime_a / attack = sustain * (1 - stateTime_r/release)
					// stateTime_a = attack * sustain * (1 - stateTime_r/release)
					//
					// given that the current value of stateTime is stateTime_r:

					stateTime = attack * sustain * (1 - stateTime/release)
				} else {
					stateTime = 0
				}
				state = nextState
			}
			// now, generate the sample
			switch state {
			case stateOff:
				samples[i] = 0
			case stateAttack:
				samples[i] = stateTime / attack
			case stateDecay:
				samples[i] = 1 - (stateTime/decay)*(1-sustain)
			case stateSustain:
				samples[i] = sustain
			case stateRelease:
				samples[i] = sustain * (1 - stateTime/release)
			}
			stateTime += 1.0 / float64(cfg.SampleRateHz)
		}

		return samples
	})
}

func NewConstantGenerator(val float64) generator.SampleGenerator {
	var buf []float64
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		if len(buf) < n {
			buf = make([]float64, n)
			for i := 0; i < n; i++ {
				buf[i] = val
			}
		}
		return buf[:n]
	})
}

func NewMultiplyGenerator() generator.SampleGenerator {
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := range res {
			res[i] = 1
			for _, s := range cfg.InputSamples {
				res[i] *= s[i]
			}
		}
		return res
	})
}

func NewAddGenerator() generator.SampleGenerator {
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := range res {
			for _, s := range cfg.InputSamples {
				res[i] += s[i]
			}
		}
		return res
	})
}

func NewDelayGenerator() generator.SampleGenerator {
	// Simulate a tape delay by using a buffer of samples with a read
	// and write pointer. If the delay is changed, we simulate a
	// physical read/write head by maintaining a sample velocity for the
	// read head. The write head is always at the end of the buffer. The
	// read head can never move backwards, so if the delay is decreased,
	// the read head will accelerate, and if the delay is increased, the
	// read head will decelerate.
	var tape []float64
	var readHead float64
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		in := cfg.InputSamples["$0"]

		// var targetDelaySecs float64
		// var targetDelaySamps float64
		// var actualDelaySamps float64
		for i := 0; i < n; i++ {
			delaySeconds := cfg.InputSamples["delay"][i]
			if delaySeconds < 0 {
				delaySeconds = 0
			}
			delaySamples := delaySeconds * float64(cfg.SampleRateHz)
			// handle the initialization case, where the tape hasn't been set up yet.
			if tape == nil {
				tape = make([]float64, int(delaySeconds*float64(cfg.SampleRateHz)))
			}
			actualDelaySamples := float64(len(tape)) - readHead

			tape = append(tape, in[i])

			if len(tape) == 1 {
				res[i] = tape[0]
			} else {
				// read the sample from the tape at the read head with linear interpolation
				// between the two adjacent samples.
				readHeadInt := int(readHead)
				readHeadFrac := readHead - float64(readHeadInt)
				res[i] = tape[readHeadInt]*(1-readHeadFrac) + tape[readHeadInt+1]*readHeadFrac
			}

			const maxStep = 2
			const minStep = 1 / maxStep

			// update the read head position with max and min bounds to prevent
			// the read head from moving backwards or infinitely forward.
			if delaySamples == 0 && actualDelaySamples > 0 {
				readHead += maxStep
			} else if actualDelaySamples > maxStep*delaySamples {
				readHead += maxStep
			} else if actualDelaySamples < minStep*delaySamples {
				readHead += minStep
			} else {
				vel := actualDelaySamples / delaySamples
				if math.IsNaN(vel) {
					readHead += maxStep
				} else {
					readHead += math.Max(minStep, math.Min(maxStep, vel))
				}
			}
			if readHead >= float64(len(tape)) {
				readHead = 0
				tape = tape[:0]
			}
			// drop samples that have already been read from the tape.
			if readHead > 1 {
				tape = tape[int(readHead):]
				readHead = readHead - math.Floor(readHead)
			}

			// targetDelaySecs = delaySeconds
			// targetDelaySamps = delaySamples
			// actualDelaySamps = actualDelaySamples
		}

		// fmt.Printf("sample diff: %v, target delay sec: %v, target delay samps: %v, actual delay samps: %v, read head: %v\n, ratio: %v\n", targetDelaySamps-actualDelaySamps, targetDelaySecs, targetDelaySamps, actualDelaySamps, readHead, actualDelaySamps/targetDelaySamps)

		return res
	})
}

func NewFreeverbGenerator() generator.SampleGenerator {
	revmod := freeverb.NewRevModel()
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		if wet := cfg.InputSamples["wet"]; len(wet) > 0 {
			revmod.SetWet(float32(wet[0]))
		}
		if damp := cfg.InputSamples["damp"]; len(damp) > 0 {
			revmod.SetDamp(float32(damp[0]))
		}
		if room := cfg.InputSamples["room"]; len(room) > 0 {
			revmod.SetRoomSize(float32(room[0]))
		}

		input32 := make([]float32, n)
		for i := 0; i < n; i++ {
			input32[i] = float32(cfg.InputSamples["$0"][i])
		}
		outputLeft := make([]float32, n)
		outputRight := make([]float32, n)
		revmod.ProcessReplace(input32, input32, outputLeft, outputRight, n, 1)
		output := make([]float64, n)
		for i := 0; i < n; i++ {
			output[i] = 0.5 * (float64(outputLeft[i]) + float64(outputRight[i]))
		}

		return output
	})
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
			//fmt.Println("XXX clipping high")
		}
		if s < 0 {
			//fmt.Println("XXX clipping low")
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

	program, err := mratlang.Parse(strings.NewReader(string(synthFile)))
	if err != nil {
		fmt.Println("error parsing script:", err)
		return
	}
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	g, sinkChannels, err := program.Eval(mratlang.WithLoadPath([]string{pwd}))
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

type value interface{}

type command struct {
	ref     string
	sigName string
	sigArgs map[string]value
}

func (c command) String() string {
	res := fmt.Sprintf("(%s", c.sigName)
	for k, v := range c.sigArgs {
		res += fmt.Sprintf(" %s=%v", k, v)
	}
	res += ")"
	return res
}

func parseCommand(cmd string) (command, error) {
	// parse the command. It should be of the form:
	// [ref = ]<sin|saw|sqr|noise> [(OSC_ARGS)]
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

	// check if the command has a ref
	if fields[1] == "=" {
		res.ref = fields[0]
		fields = fields[2:]
	}

	if len(fields) < 2 {
		return res, fmt.Errorf("invalid command: %s", cmd)
	}

	res.sigName = fields[0]
	fields = fields[1:]

	res.sigArgs = make(map[string]value)
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
			var val value
			if unicode.IsLetter([]rune(fields[2])[0]) {
				// if the value is a reference to another signal, store it as a string
				val = fields[2]
			} else {
				var err error
				val, err = strconv.ParseFloat(fields[2], 64)
				if err != nil {
					return res, fmt.Errorf("invalid command, value not a float: %s", fields[2])
				}
			}
			res.sigArgs[key] = val
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

	if len(fields) > 0 {
		return res, fmt.Errorf("invalid command: %s", cmd)
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

var (
	commentsRe = regexp.MustCompile(`\#.*$`)
)

func scriptToGraph(synthDesc string) (g *graph.Graph, sinks []<-chan []float64, err error) {
	// split synth file into lines
	lines := strings.Split(synthDesc, "\n")
	commands := make([]string, 0, len(lines))
	for _, line := range lines {
		// remove comments
		line = commentsRe.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		commands = append(commands, line)
	}

	g = &graph.Graph{}
	sinkID, graphOutputChannel := g.AddSinkNode(graph.WithLabel("output"))
	var oscNodeIDs []graph.NodeID

	nodesByRef := make(map[string]graph.NodeID)
	for _, cmd := range commands {
		command, cmdErr := parseCommand(cmd)
		if cmdErr != nil {
			return nil, nil, cmdErr
		}

		argGenerators := make(map[string]graph.NodeID)
		for k, v := range command.sigArgs {
			switch v := v.(type) {
			case float64:
				// if the argument is a float, add a constant node
				argGenerators[k] = g.AddGeneratorNode(NewConstantGenerator(v), graph.WithLabel(fmt.Sprintf("%.3f", v)))
			case string:
				// if the argument is a string, it should be a reference to
				// another signal or a note name. TODO: build midi note names
				// into the language so that this is unambiguous.
				if ref, ok := nodesByRef[v]; ok {
					argGenerators[k] = ref
				} else {
					freq, err := parseFreq(v)
					if err != nil {
						return nil, nil, fmt.Errorf("invalid command, unknown frequency: %s", v)
					}
					argGenerators[k] = g.AddGeneratorNode(NewConstantGenerator(freq), graph.WithLabel(v))
				}
			}
		}

		var gen generator.SampleGenerator

		switch command.sigName {
		case "*":
			gen = NewMultiplyGenerator()
		case "+":
			gen = NewAddGenerator()
		case "sin":
			gen = wavtabGenerator(wavtabs.Sin(1024))
		case "saw":
			gen = wavtabGenerator(wavtabs.Saw(1024))
		case "sqr":
			_, ok := command.sigArgs["dc"]
			if !ok {
				command.sigArgs["dc"] = 0.5
				argGenerators["dc"] = g.AddGeneratorNode(NewConstantGenerator(0.5), graph.WithLabel("0.5"))
			}
			phase := 0.0
			gen = generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
				dcs := cfg.InputSamples["dc"]
				ws := cfg.InputSamples["w"]
				res := make([]float64, n)

				lastDC := dcs[0]
				wavtab := wavtabs.Square(1024, lastDC)
				w := 0.0
				for i := 0; i < n; i++ {
					if dcs[i] != lastDC {
						lastDC = dcs[i]
						wavtab = wavtabs.Square(1024, lastDC)
					}
					if i < len(ws) {
						w = ws[i]
					}
					res[i] = wavtab.Lerp(phase)
					phase += w / float64(cfg.SampleRateHz)
					if phase > 1 {
						phase -= 1
					}
				}
				return res
			})
		case "noise":
			gen = Noise
		case "env":
			gen = NewADSRGenerator()
		case "delay":
			gen = NewDelayGenerator()
		case "freeverb":
			gen = NewFreeverbGenerator()
		default:
			err = fmt.Errorf("invalid signal generator: %s", command.sigName)
			return
		}
		genID := g.AddGeneratorNode(gen, graph.WithLabel(command.String()))
		for port, input := range argGenerators {
			g.AddEdge(input, genID, port)
		}

		ampVal, ok := command.sigArgs["amp"]
		if !ok {
			if command.ref != "" {
				nodesByRef[command.ref] = genID
			} else {
				oscNodeIDs = append(oscNodeIDs, genID)
			}
			continue
		}

		var ampValNodeID graph.NodeID
		switch av := ampVal.(type) {
		case string:
			nid, ok := nodesByRef[av]
			if !ok {
				err = fmt.Errorf("invalid reference: %s", av)
				return
			}
			ampValNodeID = nid
		case float64:
			ampValNodeID = g.AddGeneratorNode(NewConstantGenerator(av), graph.WithLabel(fmt.Sprintf("%.3f", av)))
		}

		ampID := g.AddGeneratorNode(generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
			res := make([]float64, n)
			for i := range res {
				res[i] = cfg.InputSamples["$0"][i] * cfg.InputSamples["$1"][i]
			}
			return res
		}), graph.WithLabel("*"))
		g.AddEdge(genID, ampID, "$0")
		g.AddEdge(ampValNodeID, ampID, "$1")

		if command.ref != "" {
			nodesByRef[command.ref] = ampID
		} else {
			oscNodeIDs = append(oscNodeIDs, ampID)
		}
	}
	mixerNodeID := g.AddGeneratorNode(generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for _, samples := range cfg.InputSamples {
			for i := 0; i < n; i++ {
				res[i] += samples[i] / float64(len(cfg.InputSamples))
			}
		}
		return res
	}), graph.WithLabel("mixer"))
	g.AddEdge(mixerNodeID, sinkID, "$0")
	for i, id := range oscNodeIDs {
		g.AddEdge(id, mixerNodeID, fmt.Sprintf("$%d", i))
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

func (a *App) GraphDot() string {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	if a.graph == nil {
		return ""
	}
	return a.graph.Dot()
}
