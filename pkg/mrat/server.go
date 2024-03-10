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

	"github.com/jfhamlin/muscrat/pkg/bufferpool"
	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/gen/gljimports"
	"github.com/jfhamlin/muscrat/pkg/graph"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
	"github.com/jfhamlin/muscrat/pkg/stdlib"
	"github.com/jfhamlin/muscrat/pkg/ugen"

	"github.com/jfhamlin/muscrat/pkg/audio"

	"github.com/glojurelang/glojure/pkg/glj"
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
	vizBufferFlushSize = 1024
)

type (
	Server struct {
		ctx context.Context

		sampleRate int

		gain       float64
		targetGain float64

		runner *graph.Runner

		// channel for raw, unprocessed output samples.
		// one []float64 per audio channel.
		outputChannel chan [][]float64

		vizSamplesBuffer []float64

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
	out := make(chan [][]float64, 1)
	return &Server{
		gain:          1,
		targetGain:    1,
		outputChannel: out,
		msgChan:       msgChan,
		runner: graph.NewRunner(ugen.SampleConfig{
			SampleRateHz: conf.SampleRate,
		}, out),
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

	go s.runner.Run(ctx)
	go s.sendSamples()

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
	s.runner.SetGraph(g)
}

func (s *Server) sendSamples() {
	timePerBuf := time.Duration(conf.BufferSize) * time.Second / time.Duration(audio.SampleRate())
	for {
		start := time.Now()
		channelSamples := <-s.outputChannel
		if len(channelSamples) < audio.NumChannels() {
			fmt.Printf("WARNING: expected %d channels, got %d\n", audio.NumChannels(), len(channelSamples))
			continue
		}
		if dur := time.Since(start); dur > timePerBuf && false { // Enable to debug buffer timeouts.
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
					pubsub.Publish("samples", *vizEmitBuffer)
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

func zeroGraph() *graph.Graph {
	return graph.SExprToGraph(glj.Read(`
		{:nodes ({:id "3", :type :out, :ctor nil, :args [0], :key nil, :sink true}
             {:id "4", :type :out, :ctor nil, :args [1], :key nil, :sink true})}
`))
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
