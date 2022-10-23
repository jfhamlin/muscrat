package dsp

import (
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
