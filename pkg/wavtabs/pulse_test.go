package wavtabs

import (
	"context"
	"testing"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func TestPulse(t *testing.T) {
	const resolution = 10
	tbl := Pulse(resolution)
	for x := 0.0; x <= 1.0; x += 0.01 {
		v := tbl.Lerp(x)
		if x < 1-(1/float64(resolution)) {
			if v != 1 {
				t.Errorf("tbl.Lerp(%f) = %f, want 1", x, v)
			}
		}
	}
	// fmt.Println(plot.LineChartString(tbl.tbl, 80, 20))
}

func TestPulseSample(t *testing.T) {
	gen := Generator(Pulse(100), WithDefaultDutyCycle(0.5))
	const sampleRate = 44100
	const numSamples = 100

	// freq gives us one full period in numSamples samples
	freq := float64(sampleRate) / float64(numSamples)

	freqSamples := make([]float64, numSamples)
	for i := 0; i < numSamples; i++ {
		freqSamples[i] = freq
	}
	samps := make([]float64, numSamples)
	gen.Gen(context.Background(), ugen.SampleConfig{
		SampleRateHz: sampleRate,
		InputSamples: map[string][]float64{"w": freqSamples},
	}, samps)

	countHigh := 0
	countLow := 0
	for _, s := range samps {
		if s > 0 {
			countHigh++
		}
		if s < 0 {
			countLow++
		}
	}
	if countHigh != numSamples/2 {
		t.Errorf("countHigh = %d, want %d", countHigh, numSamples/2)
	}
	if countLow != numSamples/2 {
		t.Errorf("countLow = %d, want %d", countLow, numSamples/2)
	}
}
