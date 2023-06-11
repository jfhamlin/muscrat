package aio

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/gordonklaus/portaudio"
	"github.com/jfhamlin/muscrat/pkg/ugen"

	// pprof
	"net/http"
	_ "net/http/pprof"
)

func init() {
	portaudio.Initialize()

	// start pprof
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	inStreamCount  int32
	inStreamMtx    sync.RWMutex
	inStreamChans  = make(map[*InputDevice]chan float64)
	inStreamSem    = make(chan struct{}, 1)
	inStreamCancel = make(chan struct{}, 1)
)

func publishStream() {
	inStreamSem <- struct{}{}
	defer func() {
		<-inStreamSem
	}()

	portaudio.Initialize()
	defer portaudio.Terminate()

	in := make([]int32, 1024)
	stream, err := portaudio.OpenDefaultStream(1, 0, 44100, len(in), in)
	chk(err)
	chk(stream.Start())
	defer func() {
		chk(stream.Stop())
	}()

	for {
		select {
		case <-inStreamCancel:
			return
		default:
			chk(stream.Read())
			inStreamMtx.RLock()
			for _, sample := range in {
				smp64 := float64(sample) / float64(math.MaxInt32)
				for _, ch := range inStreamChans {
					ch <- smp64
				}
			}
			inStreamMtx.RUnlock()
		}
	}
}

type InputDevice struct {
	sampleChan chan float64
	started    bool
}

func (in *InputDevice) Start(ctx context.Context) error {
	inStreamMtx.Lock()
	defer inStreamMtx.Unlock()
	if in.started {
		return nil
	}
	in.started = true

	if inStreamCount == 0 {
		go publishStream()
	}

	inStreamCount++
	inStreamChans[in] = in.sampleChan
	return nil
}

func (in *InputDevice) Stop(ctx context.Context) error {
	inStreamMtx.RLock()
	if !in.started {
		inStreamMtx.RUnlock()
		return nil
	}
	if inStreamCount == 1 {
		inStreamCancel <- struct{}{}
	}
	inStreamMtx.RUnlock()

	inStreamMtx.Lock()
	defer inStreamMtx.Unlock()

	inStreamCount--
	delete(inStreamChans, in)
	return nil
}

func NewInputDevice() ugen.SampleGenerator {
	return &InputDevice{
		sampleChan: make(chan float64, 1024),
	}
}

func (in *InputDevice) GenerateSamples(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
	samples := make([]float64, n)
	for i := range samples {
		samples[i] = <-in.sampleChan
	}
	return samples
}
