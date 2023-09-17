package wavtabs

import (
	"gonum.org/v1/gonum/dsp/fourier"
)

// fft computes the FFT of the wave table.
func (t *Table) fft() {
	if t.bins != nil {
		return
	}

	fft := fourier.NewFFT(t.Resolution())
	t.bins = fft.Coefficients(nil, t.tbl[:t.Resolution()])
}

// bandLimited returns a band-limited version of the wave table with
// the given number of cycles per second and frequency limit. Cycles
// per second is the number of times the wave is repeated per second,
// and frequency limit is the maximum frequency to include in the
// band-limited waveform. The returned wave table will have the same
// resolution as the original, but will omit frequencies above the
// given limit.
func (t *Table) bandLimited(cyclesPerSec, freqLimit float64) []float64 {
	t.fft()

	// treat the table as a sequence of samples
	// if the table is repeated cyclesPerSec times per second,
	// then sample rate is cyclesPerSec * t.Resolution()
	//
	// the center frequency of the nth bin is n * sampleRate / t.Resolution()
	// i.e. n * cyclesPerSec
	//
	// the max frequency of the nth bin is (n+0.5) * cyclesPerSec

	deltaFreq := cyclesPerSec

	binsCpy := make([]complex128, len(t.bins))
	copy(binsCpy, t.bins)

	for i := range binsCpy {
		binMaxFreq := (float64(i) + 0.5) * deltaFreq
		if binMaxFreq >= freqLimit {
			binsCpy[i] = 0
		}
	}

	fft := fourier.NewFFT(t.Resolution())

	blTable := fft.Sequence(nil, binsCpy)
	for i := range blTable {
		blTable[i] /= float64(t.Resolution())
	}
	return blTable
}
