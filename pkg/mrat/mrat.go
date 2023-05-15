package mrat

import (
	"bytes"
	"context"
	"crypto/sha256"
	"math"
	"os"
	"sync"
	"time"

	"github.com/bspaans/bleep/audio"
	"github.com/bspaans/bleep/sinks"

	"github.com/jfhamlin/muscrat/pkg/gen/gljimports"
	"github.com/jfhamlin/muscrat/pkg/graph"
	"github.com/jfhamlin/muscrat/pkg/ugen"

	"github.com/glojurelang/glojure/pkgmap"
	"github.com/glojurelang/glojure/runtime"
)

func init() {
	// TODO: enable setting a dynamic stdlib path vs. using the default.
	runtime.AddLoadPath(os.DirFS("./pkg/stdlib")) //stdlib.StdLib)
	runtime.AddLoadPath(os.DirFS("."))

	gljimports.RegisterImports(func(export string, val interface{}) {
		pkgmap.Set(export, val)
	})
}

type (
	Server struct {
		sampleRate int

		gain       float64
		targetGain float64

		// channel for raw, unprocessed output samples.
		// one []float64 per audio channel.
		outputChannel chan [][]float64

		graphRunner *graphRunner

		lastFileHash [32]byte

		// output channel for server messages
		msgChan chan<- *ServerMessage

		mtx sync.RWMutex
	}

	ServerMessage struct {
		Text string
	}
)

func NewServer(msgChan chan<- *ServerMessage) *Server {
	return &Server{
		gain:          0.25,
		targetGain:    0.25,
		outputChannel: make(chan [][]float64, 4),
		msgChan:       msgChan,
	}
}

func (s *Server) Start(script, path string) error {
	cfg := audio.NewAudioConfig()
	s.sampleRate = cfg.SampleRate

	// set up the audio output
	sink, err := sinks.NewSDLSink(cfg)
	if err != nil {
		panic(err)
	}
	sink.Start(s.getSamples)

	return s.EvalScript(script, path)
}

func (s *Server) EvalScript(script, path string) error {
	hash := sha256.Sum256([]byte(script))

	s.mtx.RLock()
	if bytes.Equal(hash[:], s.lastFileHash[:]) {
		s.mtx.RUnlock()
		return nil
	}
	s.mtx.RUnlock()

	g, err := EvalScript(script, path)
	if err != nil {
		return err
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.lastFileHash = hash

	s.playGraph(g)
	return nil
}

func (s *Server) playGraph(g *graph.Graph) {
	gr := s.newGraphRunner(g)
	go gr.run()

	if s.graphRunner != nil {
		s.graphRunner.outputTo(nil)
		s.fadeTo(gr)
		s.graphRunner.stop()
	}
	s.graphRunner = gr
	gr.outputTo(s.outputChannel)
}

func (s *Server) fadeTo(gr *graphRunner) {
	if s.graphRunner == nil {
		return
	}

	ogr := s.graphRunner
	ngr := gr

	const fadeTimeMS = 100
	fadeSamples := fadeTimeMS * s.sampleRate / 1000

	sampleIndex := 0
	for sampleIndex < fadeSamples {
		oldSmps := <-ogr.graphOutputCh
		newSmps := <-ngr.graphOutputCh

		// mixed samples
		mixSmps := make([][]float64, int(math.Max(float64(len(oldSmps)), float64(len(newSmps)))))

		zeros := make([]float64, len(oldSmps[0]))

		for chIdx := 0; chIdx < len(mixSmps); chIdx++ {
			var os []float64
			var ns []float64
			if chIdx < len(oldSmps) {
				os = oldSmps[chIdx]
			} else {
				os = zeros
			}
			if chIdx < len(newSmps) {
				ns = newSmps[chIdx]
			} else {
				ns = zeros
			}

			mixSmps[chIdx] = make([]float64, len(os))
			for i := range os {
				oldS := os[i]
				newS := ns[i]
				t := float64(sampleIndex) / float64(fadeSamples)
				if t > 1 {
					t = 1
				}
				mixSmps[chIdx][i] = oldS*(1-t) + newS*t
				sampleIndex++
			}
		}
		s.outputChannel <- mixSmps
	}
}

func (s *Server) getSamples(cfg *audio.AudioConfig, n int) []int {
	var channelSamples [][]float64
	select {
	case channelSamples = <-s.outputChannel:
	case <-time.After(time.Duration(n) * time.Second / time.Duration(cfg.SampleRate)):
		// return silence if we can't get samples fast enough.
		channelSamples = [][]float64{make([]float64, n)}
	}

	// update gain to approach target gain.
	for i, samples := range channelSamples {
		newSamples := make([]float64, len(samples))
		target := s.targetGain
		gainStep := (target - s.gain) / float64(len(samples))
		for i, smp := range samples {
			newSamples[i] = smp * s.gain
			s.gain += gainStep
		}
		s.gain = target
		channelSamples[i] = newSamples
	}

	return transformSampleBuffer(cfg, channelSamples)
}

////////////////////////////////////////////////////////////////////////////////

type (
	graphRunner struct {
		sampleRate    int
		graph         *graph.Graph
		graphOutputCh chan [][]float64

		cancel context.CancelFunc

		cancelOutputTo func()

		stopped chan struct{}
	}
)

func (s *Server) newGraphRunner(g *graph.Graph) *graphRunner {
	return &graphRunner{
		sampleRate:    s.sampleRate,
		graph:         g,
		graphOutputCh: make(chan [][]float64),
		stopped:       make(chan struct{}),
	}
}

func (gr *graphRunner) run() {
	ctx, cancel := context.WithCancel(context.Background())
	gr.cancel = cancel

	go func() {
		defer close(gr.stopped)
		graph.RunGraph(ctx, gr.graph, ugen.SampleConfig{SampleRateHz: gr.sampleRate})
	}()

	defer close(gr.graphOutputCh)
	for {
		output := make([][]float64, 0, 2)
		for _, sinkChan := range gr.graph.SinkChans() {
			select {
			case samps := <-sinkChan:
				output = append(output, samps)
			case <-ctx.Done():
				return
			}
		}
		gr.graphOutputCh <- output
	}
}

func (gr *graphRunner) stop() {
	gr.cancel()
	<-gr.stopped
}

func (gr *graphRunner) outputTo(output chan<- [][]float64) {
	if output == nil {
		if gr.cancelOutputTo != nil {
			gr.cancelOutputTo()
		}
		gr.cancelOutputTo = nil
		return
	}
	stopChan := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(stopChan)
		for {
			select {
			case <-ctx.Done():
				return
			case sampleChannels, ok := <-gr.graphOutputCh:
				if !ok {
					return
				}
				output <- sampleChannels
			}
		}
	}()
	gr.cancelOutputTo = func() {
		cancel()
		<-stopChan
	}
}

////////////////////////////////////////////////////////////////////////////////

func transformSampleBuffer(cfg *audio.AudioConfig, buf [][]float64) []int {
	var out []int
	if cfg.Stereo {
		out = make([]int, 2*len(buf[0]))
	} else {
		out = make([]int, len(buf[0]))
	}

	maxValue := float64(int(1) << cfg.BitDepth)
	transformSample := func(sample float64) int {
		s := (sample + 1) * (maxValue / 2)
		if s > maxValue {
			//fmt.Printf("XXX clipping high (max=%v): %v (%v)\n", maxValue, s, sample)
		}
		if s < 0 {
			//fmt.Printf("XXX clipping low (min=%v): %v (%v)\n", 0, s, sample)
		}
		s = math.Max(0, math.Ceil(s))
		return int(math.Min(s, maxValue-1))
	}

	zeroSample := transformSample(0)
	for i := range buf[0] {
		if cfg.Stereo {
			out[2*i] = transformSample(buf[0][i])
			if len(buf) > 1 {
				out[2*i+1] = transformSample(buf[1][i])
			} else {
				out[2*i+1] = zeroSample
			}
		} else {
			out[i] = transformSample(buf[0][i])
		}
	}

	return out
}