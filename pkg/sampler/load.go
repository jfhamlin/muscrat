package sampler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"
	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/mewkiz/flac"
)

func LoadSample(filename string) (res []float64) {
	if data, ok := cache.get(filename); ok {
		return data
	}

	defer func() {
		cache.set(filename, res)
	}()

	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fileType := filepath.Ext(filename)
	switch fileType {
	case ".wav":
		return loadWav(filename, f)
	case ".flac":
		return loadFlac(filename, f)
	case ".mp3":
		return loadMP3(filename, f)
	default:
		panic(fmt.Errorf("load-sample: unsupported file type: %s", fileType))
	}
}

// TODO: multiple channels

func loadFlac(filename string, f *os.File) []float64 {
	stream, err := flac.New(f)
	if err != nil {
		panic(fmt.Errorf("load-sample: error parsing FLAC header: %v", err))
	}

	// Determine the scaling factor based on bit depth
	maxVal := 1 << (stream.Info.BitsPerSample - 1)
	scaleFactor := 1.0 / float64(maxVal)

	var samples []float64
	for {
		frame, err := stream.ParseNext()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(fmt.Errorf("load-sample: error parsing FLAC data: %v", err))
		}
		for i := 0; i < len(frame.Subframes[0].Samples); i++ {
			var sum float64
			for _, block := range frame.Subframes {
				sum += float64(block.Samples[i])
			}
			avg := sum / float64(len(frame.Subframes))
			samples = append(samples, avg*scaleFactor)
		}
	}

	// if stream.Info.SampleRate != sampleRate {
	// 	rsmp := resampler.NewWithSkipZeros(1, micSampleRate, conf.SampleRate, 10)
	// 	samples, err = rsmp.Process(samples)
	// }

	return samples
}

func loadWav(filename string, f *os.File) []float64 {
	dec := wav.NewDecoder(f)
	if !dec.IsValidFile() {
		panic(fmt.Errorf("load-sample: file '%s' is not a valid WAV file", filename))
	}

	var intSamples []int
	audioBuf := &audio.IntBuffer{Data: make([]int, 2048)}
	for {
		n, err := dec.PCMBuffer(audioBuf)
		if err != nil {
			panic(fmt.Errorf("load-sample: error reading PCM data: %v", err))
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

func loadMP3(filename string, f *os.File) []float64 {
	dec, err := mp3.NewDecoder(f)
	if err != nil {
		panic(fmt.Errorf("load-sample: error creating MP3 decoder: %v", err))
	}

	sampleRate := dec.SampleRate()

	var intSamples []int
	for {
		data := make([]byte, 2048)
		n, err := dec.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(fmt.Errorf("load-sample: error reading MP3 data: %v", err))
		}
		// a sample is always 4 bytes (two 16-bit samples, little-endian,
		// one for each channel) we average the two channels to get a
		// single sample
		for i := 0; i < n; i += 4 {
			sample1 := int(int16(data[i]) | int16(data[i+1])<<8)
			sample2 := int(int16(data[i+2]) | int16(data[i+3])<<8)
			intSamples = append(intSamples, (sample1+sample2)/2)
		}
	}

	floatSamples := make([]float64, 0, len(intSamples))
	for _, s := range intSamples {
		floatSample := float64(s) / float64(1<<15)
		if floatSample > 1 {
			floatSample = 1
		} else if floatSample < -1 {
			floatSample = -1
		}
		floatSamples = append(floatSamples, floatSample)
	}

	if sampleRate != conf.SampleRate {
		outputSamples := make([]float64, len(floatSamples)*conf.SampleRate/int(sampleRate))
		for i := range outputSamples {
			t := float64(i) / float64(len(outputSamples)-1)
			outputSamples[i] = floatSamples[int(t*float64(len(floatSamples)-1))]
		}
		floatSamples = outputSamples
	}

	return floatSamples
}
