package audio

import (
	"fmt"
	"unsafe"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
	"github.com/veandco/go-sdl2/sdl"
)

func init() {
	// initialize SDL audio
	if err := sdl.InitSubSystem(sdl.INIT_AUDIO); err != nil {
		panic(err)
	}
}

type (
	Option func(*options)

	options struct{}
)

const (
	maxQueuedBuffers = 1
	sampleSize       = 4 // assume 32-bit float samples
)

var (
	float32Buf []float32

	audioSpec *sdl.AudioSpec
)

func Open(opts ...Option) error {
	if audioSpec != nil {
		panic("audio already open")
	}

	// By default, 32-bit float samples, stereo, sample rate
	spec := &sdl.AudioSpec{
		Freq:     int32(conf.SampleRate),
		Format:   sdl.AUDIO_F32SYS,
		Channels: 2,
		Samples:  uint16(conf.OutputBufferSize),
	}
	if err := sdl.OpenAudio(spec, nil); err != nil {
		return err
	}
	sdl.PauseAudio(false)
	audioSpec = spec

	float32Buf = make([]float32, int(spec.Samples)*int(spec.Channels))

	return nil
}

func Close() {
	sdl.CloseAudio()
	audioSpec = nil
}

func SampleRate() int {
	return int(audioSpec.Freq)
}

func NumChannels() int {
	return int(audioSpec.Channels)
}

func numBytesToNumSamples(numBytes uint32) int {
	return int(numBytes) / int(audioSpec.Channels*sampleSize)
}

// var (
// 	timeOfLastQueuedAudio = time.Now()
// )

func QueueAudioFloat64(fbuf []float64) error {
	// if dur := time.Since(timeOfLastQueuedAudio); dur > time.Millisecond {
	// 	fmt.Printf("time since last queued audio: %v\n", dur)
	// }
	// defer func() {
	// 	timeOfLastQueuedAudio = time.Now()
	// }()

	bufferByteSize := sampleSize * int(audioSpec.Channels) * int(audioSpec.Samples)
	sz := sdl.GetQueuedAudioSize(1)
	if sz < 1024 {
		fmt.Println(fmt.Sprintf("audio buffer underflow: %d < %d", sz, bufferByteSize))
		pubsub.Publish("console.debug", fmt.Sprintf("audio buffer underflow: %d < %d", sz, bufferByteSize))
	}
	for numBytesToNumSamples(sdl.GetQueuedAudioSize(1)) > maxQueuedBuffers*bufferByteSize {
	}

	for i, f := range fbuf {
		float32Buf[i] = float32(f)
	}
	sendBuf := float32Buf[:len(fbuf)]
	buf := unsafe.Slice((*byte)(unsafe.Pointer(&sendBuf[0])), len(sendBuf)*4)
	return sdl.QueueAudio(1, buf)
}
