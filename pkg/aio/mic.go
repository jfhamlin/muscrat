package aio

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
	"github.com/jfhamlin/muscrat/pkg/ugen"
	"github.com/oov/audio/resampler"

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

const (
	bufDur = 20 * time.Millisecond // grab 20ms of audio at a time
)

var (
	inStreamCount  int32
	inStreamMtx    sync.RWMutex
	inStreamChans  = make(map[*InputDevice]chan []float64)
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

	micSampleRate := conf.SampleRate
	if !sampleRateSupported(micSampleRate) {
		// fall back to 44100, which is very likely to be supported on all
		// systems.
		micSampleRate = 44100
	}
	bufSize := int(float64(micSampleRate) * bufDur.Seconds())

	// :shrug: just use quality 10 (0-10)
	rsmp := resampler.NewWithSkipZeros(1, micSampleRate, conf.SampleRate, 10)
	ratio := float64(conf.SampleRate) / float64(micSampleRate)

	resampleOutBuf := make([]float64, int(float64(bufSize)*ratio+1))
	in := make([]int32, bufSize)

	stream, err := portaudio.OpenDefaultStream(1, 0, float64(micSampleRate), len(in), in)
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

			buf := make([]float64, len(in))
			for i, sample := range in {
				smp64 := float64(sample) / float64(math.MaxInt32)
				buf[i] = smp64
			}
			out := buf
			if micSampleRate != conf.SampleRate {
				rn, wn := rsmp.ProcessFloat64(0, buf, resampleOutBuf)
				if rn != len(buf) {
					panic(fmt.Sprintf("resampler did not process all samples: %d != %d", rn, len(buf)))
				}
				out = resampleOutBuf[:wn]
			}
			for _, ch := range inStreamChans {
				select {
				case ch <- out:
				default: // don't wait for slow consumers
					pubsub.Publish("console.debug", "dropping audio sample")
				}
			}
			inStreamMtx.RUnlock()
		}
	}
}

func sampleRateSupported(rate int) bool {
	defaultInputDevice, err := portaudio.DefaultInputDevice()
	if err != nil {
		return false
	}

	err = portaudio.IsFormatSupported(portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   defaultInputDevice,
			Channels: 1,
		},
		SampleRate: float64(rate),
	}, make([]int32, 1024))

	return err == nil
}

type InputDevice struct {
	latestBuf  []float64
	sampleChan chan []float64
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

func NewInputDevice() ugen.UGen {
	return &InputDevice{
		// buffering to avoid blocking on intermittent interruptions
		sampleChan: make(chan []float64, 4),
	}
}

func (in *InputDevice) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := 0; i < len(out); i++ {
		if len(in.latestBuf) == 0 {
			in.latestBuf = <-in.sampleChan
		}

		out[i] = in.latestBuf[0]
		in.latestBuf = in.latestBuf[1:]
	}
}
