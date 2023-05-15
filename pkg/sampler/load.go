package sampler

import (
	"fmt"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
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

	floatSamples := make([]float64, len(intSamples))

	for _, s := range intSamples {
		floatSample := float64(s) / float64(int(1)<<uint(bitDepth-1))
		if floatSample > 1 {
			floatSample = 1
		} else if floatSample < -1 {
			floatSample = -1
		}
		floatSamples = append(floatSamples, floatSample)
	}

	// TODO: we're getting a huge prefix of zero samples that don't play
	// when playing the wav file with Quicktime. We may be loading a
	// channel with no audio?
	for i, s := range floatSamples {
		if s != 0 {
			floatSamples = floatSamples[i:]
			break
		}
	}

	// resample to 44100 Hz, assumed to be the sample rate of the audio device
	// TODOs:
	// - make this configurable
	// - don't assume 44100 Hz
	const deviceSampleRate = 44100
	if dec.SampleRate != deviceSampleRate {
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
