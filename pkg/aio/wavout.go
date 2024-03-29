package aio

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

type WavOut struct {
	enc   *wav.Encoder
	f     *os.File
	fname string
}

var (
	_ ugen.UGen = (*WavOut)(nil)
)

func NewWavOut(fname string) *WavOut {
	return &WavOut{
		fname: fname,
	}
}

func (wo *WavOut) Start(ctx context.Context) error {
	var err error
	wo.f, err = os.Create(wo.fname)
	if err != nil {
		return err
	}
	return nil
}

func (wo *WavOut) Stop(ctx context.Context) error {
	if wo.enc == nil {
		return fmt.Errorf("WavOut not started")
	}
	if err := wo.enc.Close(); err != nil {
		return err
	}
	if err := wo.f.Close(); err != nil {
		return err
	}
	return nil
}

func (wo *WavOut) Gen(ctx context.Context, cfg ugen.SampleConfig, _ []float64) {
	ch0 := cfg.InputSamples["$0"]
	ch1 := cfg.InputSamples["$1"]
	if len(ch0) == 0 {
		return
	}

	numChan := 2
	if len(ch1) == 0 {
		numChan = 1
	}

	if wo.enc == nil {
		wo.enc = wav.NewEncoder(wo.f, cfg.SampleRateHz, 32, numChan, 1)
	}

	n := len(ch0)

	buf := &audio.IntBuffer{
		Format: &audio.Format{
			SampleRate:  cfg.SampleRateHz,
			NumChannels: numChan,
		},
		Data: make([]int, n*numChan),
	}

	for i := 0; i < n; i++ {
		buf.Data[i*numChan] = float64ToInt32(ch0[i])
		if numChan == 2 {
			buf.Data[i*numChan+1] = float64ToInt32(ch1[i])
		}
	}
	wo.enc.Write(buf)
}

func float64ToInt32(f float64) int {
	if f > 1.0 {
		fmt.Printf("clipping: %f > 1.0\n", f)
		f = 1.0
	}
	if f < -1.0 {
		fmt.Printf("clipping: %f < -1.0\n", f)
		f = -1.0
	}
	res := int(f * math.MaxInt32)
	return res
}
