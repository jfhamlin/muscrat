package aio

import (
	"context"
	"fmt"
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

func NewWavOut(fname string) *WavOut {
	return &WavOut{
		fname: fname,
	}
}

func (wo *WavOut) Start() error {
	var err error
	wo.f, err = os.Create(wo.fname)
	if err != nil {
		return err
	}
	return nil
}

func (wo *WavOut) Stop() error {
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

func (wo *WavOut) GenerateSamples(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
	ch0 := cfg.InputSamples["$0"]
	ch1 := cfg.InputSamples["$1"]
	if len(ch0) == 0 {
		return make([]float64, n)
	}

	numChan := 2
	if len(ch1) == 0 {
		numChan = 1
	}

	if wo.enc == nil {
		wo.enc = wav.NewEncoder(wo.f, cfg.SampleRateHz, 24, numChan, 1)
	}

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

	return make([]float64, n)
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
	return int(f * (1<<31 - 1))
}
