package aio

import (
	"context"
	"fmt"

	"github.com/MarkKremer/microphone"
	"github.com/faiface/beep"
	"github.com/gordonklaus/portaudio"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func init() {
	microphone.Init()
}

func NewMicrophone() ugen.SampleGenerator {
	dev, err := portaudio.DefaultInputDevice()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Default input device: %+v\n", *dev)

	stream, format, err := microphone.OpenDefaultStream(beep.SampleRate(int(dev.DefaultSampleRate)), 1)
	if err != nil {
		panic(err)
	}
	_ = format
	sampleChan := make(chan float64, 1024)
	go func() {
		for {
			samples := make([][2]float64, 1024)
			cnt, ok := stream.Stream(samples)
			if ok {
				for i := 0; i < cnt; i++ {
					sampleChan <- samples[i][0]
				}
			}
		}
	}()

	return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		samples := make([]float64, n)
		for i := range samples {
			samples[i] = <-sampleChan
		}
		return samples
	})
}
