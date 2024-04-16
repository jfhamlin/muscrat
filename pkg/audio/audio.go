package audio

import (
	"sync"

	"github.com/jfhamlin/muscrat/pkg/conf"
)

type (
	Option func(*options)

	options struct{}
)

const (
	maxQueuedBuffers = 1

	numChannels = 2
	sampleRate  = 44100
)

var (
	myCtx *context

	// safetyClipThreshold is the maximum value that can be queued to the audio
	// modeled after the safety clip threshold in supercollider.
	safetyClipThreshold = 1.0
)

func SampleRate() int {
	return sampleRate
}

func NumChannels() int {
	return numChannels
}

func Open(opts ...Option) error {
	ctx, err := newContext(SampleRate(), NumChannels(), 4*numChannels*conf.OutputBufferSize)
	if err != nil {
		return err
	}

	myCtx = ctx

	return nil
}

func Close() {

}

var (
	pool = &sync.Pool{
		New: func() interface{} {
			return make([]float32, numChannels*conf.OutputBufferSize)
		},
	}
)

func QueueAudioFloat64(fbuf []float64) error {
	if myCtx == nil {
		panic("audio not open")
	}

	buf := pool.Get().([]float32)
	for i := 0; i < len(fbuf); i++ {
		val := fbuf[i]
		if val > safetyClipThreshold {
			val = safetyClipThreshold
		} else if val < -safetyClipThreshold {
			val = -safetyClipThreshold
		}
		buf[i] = float32(val)
	}

	myCtx.input <- buf

	return nil
}
