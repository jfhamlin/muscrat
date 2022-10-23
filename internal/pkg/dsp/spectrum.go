package dsp

import (
	"math"
	"math/bits"
	"strconv"
)

// FFTFreqs returns the frequencies of the FFT bins. Only the
// frequencies of bins (n/2)+1 are returned, i.e. from 0 to the
// Nyquist frequency. The frequencies represent the minimum frequency
// of each bin.
func FFTFreqs(sampleRate float64, bins int) []float64 {
	if bits.OnesCount(uint(bins)) != 1 {
		panic("bins must be a power of 2, was " + strconv.Itoa(bins))
	}
	freqs := make([]float64, bins/2+1)
	for i := range freqs {
		freqs[i] = float64(i) * sampleRate / float64(bins)
	}
	return freqs
}

// LogRange returns a logarithmically spaced range of values. The
// range is inclusive of min and max. min and max cannot have
// different signs.
func LogRange(min, max float64, n int) []float64 {
	if min*max < 0 {
		panic("min and max must have the same sign")
	}
	if min == 0 {
		panic("min cannot be 0")
	}
	if min < 0 {
		min, max = -max, -min
	}
	logMin, logMax := math.Log2(min), math.Log2(max)
	logRange := logMax - logMin
	step := logRange / float64(n)
	vals := make([]float64, n)
	for i := range vals {
		vals[i] = math.Pow(2, logMin+float64(i)*step)
	}
	return vals
}
