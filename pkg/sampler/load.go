package sampler

import (
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/jfhamlin/muscrat/pkg/conf"
)

func LoadSample(filename string) []float64 {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := wav.NewDecoder(f)
	if !dec.IsValidFile() {
		panic(fmt.Errorf("open-file: file '%s' is not a valid WAV file", filename))
	}

	var intSamples []int
	audioBuf := &audio.IntBuffer{Data: make([]int, 2048)}
	for {
		n, err := dec.PCMBuffer(audioBuf)
		if err != nil {
			panic(fmt.Errorf("open-file: error reading PCM data: %v", err))
		}
		if n == 0 {
			break
		}
		intSamples = append(intSamples, audioBuf.Data...)
	}
	bitDepth := dec.SampleBitDepth()

	floatSamples := make([]float64, 0, len(intSamples))

	for _, s := range intSamples {
		floatSample := float64(s) / float64(int(1)<<uint(bitDepth-1))
		if floatSample > 1 {
			floatSample = 1
		} else if floatSample < -1 {
			floatSample = -1
		}
		floatSamples = append(floatSamples, floatSample)
	}

	deviceSampleRate := conf.SampleRate
	// TODO: use lib to resample
	if dec.SampleRate != uint32(deviceSampleRate) {
		outputSamples := make([]float64, len(floatSamples)*deviceSampleRate/int(dec.SampleRate))
		for i := range outputSamples {
			t := float64(i) / float64(len(outputSamples)-1)
			outputSamples[i] = floatSamples[int(t*float64(len(floatSamples)-1))]
		}
		floatSamples = outputSamples
	}
	fmt.Println("loaded", filename, "with", len(floatSamples), "samples")

	return floatSamples
}
