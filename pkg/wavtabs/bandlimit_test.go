package wavtabs

import (
	"fmt"
	"math"
	"testing"

	"gonum.org/v1/gonum/dsp/fourier"

	"github.com/jfhamlin/muscrat/internal/pkg/plot"
)

func TestBandLimit(t *testing.T) {
	tbl := Saw(DefaultResolution)

	for f := 20.01; f < maxBandLimitedFreq; f *= math.Sqrt2 {
		idx, off := tbl.blTableIndexOffset(f)
		if false {
			t.Errorf("%v: %v, %v", f, idx, off)
		}
	}

	// for f := 10.0; f < 10000; f += 100 {
	// 	plotBL(t, f)
	// }
}

func plotBL(t *testing.T, cyclesPerSecond float64) {
	const (
		sampleRate = 44100
		nyquist    = sampleRate / 2
	)

	tbl := Saw(DefaultResolution)
	bl := tbl.bandLimited(cyclesPerSecond, nyquist)

	sampleTable := New(bl)

	samples := make([]float64, 1024)
	for i := range samples {
		samples[i] = sampleTable.Hermite(cyclesPerSecond * float64(i) / sampleRate)
	}

	fmt.Println("=====", cyclesPerSecond, "=====")
	fmt.Println(plot.LineChartString(samples, 180, 40))

	blFFT := fourier.NewFFT(len(samples)).Coefficients(nil, samples)
	// blFFT only has len(bl)/2+1 bins; add the mirrored bins
	for i := len(blFFT) - 2; i > 0; i-- {
		blFFT = append(blFFT, blFFT[i])
	}

	fmt.Println(plot.DFTHistogramString(blFFT, sampleRate, 180, 40))
}
