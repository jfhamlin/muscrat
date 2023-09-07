package mrat

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	wrt "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/jfhamlin/muscrat/pkg/bufferpool"
	"github.com/jfhamlin/muscrat/pkg/gen/gljimports"
	"github.com/jfhamlin/muscrat/pkg/graph"
	"github.com/jfhamlin/muscrat/pkg/stdlib"
	"github.com/jfhamlin/muscrat/pkg/ugen"

	"github.com/jfhamlin/muscrat/pkg/audio"

	"github.com/glojurelang/glojure/pkg/pkgmap"
	"github.com/glojurelang/glojure/pkg/runtime"
)

func init() {
	// TODO: enable setting a dynamic stdlib path vs. using the default.
	if os.Getenv("MUSCRAT_STDLIB_PATH") == "" {
		runtime.AddLoadPath(stdlib.StdLib)
	} else {
		runtime.AddLoadPath(os.DirFS(os.Getenv("MUSCRAT_STDLIB_PATH")))
	}

	runtime.AddLoadPath(os.DirFS("."))

	gljimports.RegisterImports(func(export string, val interface{}) {
		pkgmap.Set(export, val)
	})
}

const (
	bufferSize = 128

	vizBufferFlushSize = 1024
)

type (
	Server struct {
		// wails context
		ctx context.Context

		sampleRate int

		gain       float64
		targetGain float64

		// channel for raw, unprocessed output samples.
		// one []float64 per audio channel.
		outputChannel chan [][]float64

		vizSamplesBuffer []float64

		graphRunner *graphRunner

		lastFileHash [32]byte

		// output channel for server messages
		msgChan chan<- *ServerMessage

		started bool

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
		outputChannel: make(chan [][]float64, 1),
		msgChan:       msgChan,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if s.started {
		return fmt.Errorf("server already started")
	}
	s.started = true

	s.ctx = ctx

	if err := audio.Open(); err != nil {
		return err
	}
	s.sampleRate = audio.SampleRate()

	go s.sendSamples()

	// cfg := audio.NewAudioConfig()
	// s.sampleRate = cfg.SampleRate

	// // set up the audio output
	// sink, err := sinks.NewSDLSink(cfg)
	// if err != nil {
	// 	panic(err)
	// }
	// sink.Start(s.getSamples)

	s.playGraph(zeroGraph())
	return nil
}

func (s *Server) EvalScript(path string) error {
	script, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(script)

	s.mtx.RLock()
	if bytes.Equal(hash[:], s.lastFileHash[:]) {
		s.mtx.RUnlock()
		return nil
	}
	s.mtx.RUnlock()

	g, err := EvalScript(path)
	if err != nil {
		return err
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.lastFileHash = hash

	s.playGraph(g)
	return nil
}

func (s *Server) SetGain(gain float64) {
	s.targetGain = math.Max(0, math.Min(gain, 1))
}

////////////////////////////////////////////////////////////////////////////////

func (s *Server) playGraph(g *graph.Graph) {
	gr := s.newGraphRunner(s.ctx, g)
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

func (s *Server) sendSamples() {
	timePerBuf := time.Duration(bufferSize) * time.Second / time.Duration(audio.SampleRate())
	for {
		start := time.Now()
		channelSamples := <-s.outputChannel
		if len(channelSamples) < audio.NumChannels() {
			fmt.Printf("WARNING: expected %d channels, got %d\n", audio.NumChannels(), len(channelSamples))
			continue
		}
		if dur := time.Since(start); dur > timePerBuf {
			fmt.Printf("WARNING: buffer took %s to fill, expected %s\n", dur, timePerBuf)
		}

		// update gain to approach target gain.
		for i, samples := range channelSamples {
			newSamples := bufferpool.Get(len(samples))
			target := s.targetGain
			gainStep := (target - s.gain) / float64(len(samples))
			for i, smp := range samples {
				(*newSamples)[i] = smp * s.gain
				s.gain += gainStep
			}
			s.gain = target
			channelSamples[i] = *newSamples
		}

		// send samples to audio output
		{
			out := bufferpool.Get(2 * len(channelSamples[0]))
			for i := range channelSamples[0] {
				(*out)[i*2] = channelSamples[0][i]
				(*out)[i*2+1] = channelSamples[1][i]
			}
			audio.QueueAudioFloat64(*out)
			bufferpool.Put(out)
		}

		// send samples to viewer
		{
			avgSamples := bufferpool.Get(len(channelSamples[0]))
			averageBuffers(*avgSamples, channelSamples)
			s.vizSamplesBuffer = append(s.vizSamplesBuffer, (*avgSamples)...)
			if len(s.vizSamplesBuffer) > vizBufferFlushSize {
				vizEmitBuffer := bufferpool.Get(vizBufferFlushSize)
				copy(*vizEmitBuffer, s.vizSamplesBuffer[:vizBufferFlushSize])
				s.vizSamplesBuffer = s.vizSamplesBuffer[vizBufferFlushSize:]
				go func() {
					defer bufferpool.Put(vizEmitBuffer)
					wrt.EventsEmit(s.ctx, "samples", *vizEmitBuffer)
				}()
			}
		}

		for _, smps := range channelSamples {
			smps := smps
			bufferpool.Put(&smps)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type (
	graphRunner struct {
		ctx context.Context

		sampleRate    int
		graph         *graph.Graph
		graphOutputCh chan [][]float64

		cancel context.CancelFunc

		cancelOutputTo func()

		stopped chan struct{}
	}
)

func (s *Server) newGraphRunner(ctx context.Context, g *graph.Graph) *graphRunner {
	return &graphRunner{
		ctx:           ctx,
		sampleRate:    s.sampleRate,
		graph:         g,
		graphOutputCh: make(chan [][]float64),
		stopped:       make(chan struct{}),
	}
}

func (gr *graphRunner) run() {
	ctx, cancel := context.WithCancel(gr.ctx)
	gr.cancel = cancel

	go func() {
		defer close(gr.stopped)
		gr.graph.RunWorkers(ctx, ugen.SampleConfig{SampleRateHz: gr.sampleRate})
	}()

	defer close(gr.graphOutputCh)

	for {
		start := time.Now()
		output := make([][]float64, 0, 2)
		for _, sinkChan := range gr.graph.SinkChans() {
			select {
			case samps := <-sinkChan:
				output = append(output, samps)
			case <-ctx.Done():
				return
			}
		}
		dur := time.Since(start)
		if dur > 2*time.Millisecond {
			fmt.Printf("took %s to get samples from graph\n", dur)
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

func zeroGraph() *graph.Graph {
	g := &graph.Graph{BufferSize: bufferSize}
	s0 := g.AddSinkNode()
	s1 := g.AddSinkNode()
	zero := g.AddGeneratorNode(ugen.NewConstant(0))
	g.AddEdge(zero.ID(), s0.ID(), "in")
	g.AddEdge(zero.ID(), s1.ID(), "in")
	return g
}

// averageBuffers returns the average of the N buffers. If the buffers
// are not the same length, the result is undefined, and may panic.
func averageBuffers(out []float64, bufs [][]float64) {
	if len(bufs) == 1 {
		copy(out, bufs[0])
		return
	}

	for i := 0; i < len(bufs); i++ {
		for j := 0; j < len(bufs[0]); j++ {
			out[j] += bufs[i][j]
		}
	}
	for i := 0; i < len(out); i++ {
		out[i] /= float64(len(bufs))
	}
}
