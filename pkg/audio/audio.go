package audio

import (
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/jfhamlin/muscrat/pkg/conf"
)

type (
	Option func(*options)

	options struct{}

	bufferReader struct {
		// circular buffer
		buf   []byte
		pos   int
		count int
		input chan []byte
	}
)

const (
	maxQueuedBuffers = 1

	numChannels = 2
	sampleRate  = 44100
)

var (
	otoCtx *oto.Context
	player *oto.Player
	reader *bufferReader
)

func SampleRate() int {
	return sampleRate
}

func NumChannels() int {
	return numChannels
}

func Open(opts ...Option) error {
	if otoCtx != nil {
		panic("audio already open")
	}

	//bufferDur := 4 * time.Duration(float64(time.Second)*float64(conf.OutputBufferSize)/sampleRate)
	bufferDur := time.Duration(0)
	ctx, ready, err := oto.NewContext(&oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: numChannels,
		Format:       oto.FormatFloat32LE,
		BufferSize:   bufferDur,
	})
	if err != nil {
		return err
	}
	<-ready

	otoCtx = ctx
	reader = &bufferReader{
		// enough space for two output buffers
		buf:   make([]byte, 2*4*conf.OutputBufferSize*numChannels),
		input: make(chan []byte, maxQueuedBuffers),
	}
	player = otoCtx.NewPlayer(reader)
	player.SetBufferSize(4 * numChannels * conf.OutputBufferSize)

	player.Play()

	return nil
}

func Close() {
	if player != nil {
		player.Close()
		player = nil
		otoCtx = nil
		reader = nil
	}
}

var (
	pool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 4*numChannels*conf.OutputBufferSize)
		},
	}

	val = float32(-1)
	cnt = 0
)

func QueueAudioFloat64(fbuf []float64) error {
	if reader == nil {
		panic("audio not open")
	}

	buf := pool.Get().([]byte)
	if len(buf) != 4*len(fbuf) {
		panic("unexpected buffer size")
	}

	for i, flt := range fbuf {
		_ = flt
		//f32 := float32(flt)
		f32 := val
		cnt++
		if cnt > 100 {
			val = -val
			cnt = 0
		}

		floatBits := math.Float32bits(f32)
		binary.LittleEndian.PutUint32(buf[4*i:], floatBits)
	}

	reader.input <- buf

	return nil
}

func (r *bufferReader) Read(p []byte) (int, error) {
	if r.count == 0 {
		r.fillBuffer()
	}
	if len(p) > r.count {
		p = p[:r.count]
	}
	fmt.Println("reading", len(p), "bytes")

	n := 0
	for n < len(p) {
		n = copy(p, r.buf[r.pos:])
		r.pos = (r.pos + n) % len(r.buf)
	}
	r.count -= len(p)

	return len(p), nil
}

func (r *bufferReader) fillBuffer() {
	buf := <-r.input

	if (len(r.buf)-r.pos)%4 != 0 {
		panic("invalid buffer position")
	}

	insertPos := (r.pos + r.count) % len(r.buf)
	for i := 0; i < len(buf); i += 4 {
		r.buf[insertPos] = buf[i]
		r.buf[insertPos+1] = buf[i+1]
		r.buf[insertPos+2] = buf[i+2]
		r.buf[insertPos+3] = buf[i+3]
		insertPos = (insertPos + 4) % len(r.buf)
	}
	r.count += len(buf)

	pool.Put(buf)
}
