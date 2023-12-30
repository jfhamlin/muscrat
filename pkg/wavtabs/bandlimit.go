package wavtabs

import (
	"math"

	"gonum.org/v1/gonum/dsp/fourier"
)

const (
	minBandLimitedFreq       = 20
	maxBandLimitedFreq       = 20480
	bandLimitedSemitoneRange = 4

	maxAudibleFrequency = 20000
)

func (t *Table) HermiteBL(freq, x float64) float64 {
	// find the band-limited table for the greatest frequency less than
	// or equal to freq. linearly interpolate between this table and the
	// next table. if freq is greater than the greatest frequency, use
	// the greatest frequency table. if freq is less than the least
	// frequency, use the least frequency table.
	if freq <= minBandLimitedFreq {
		return t.bandLimitedTbls[0].Hermite(x)
	}
	if freq >= maxBandLimitedFreq {
		return t.bandLimitedTbls[len(t.bandLimitedTbls)-1].Hermite(x)
	}
	index, offset := t.blTableIndexOffset(freq)
	v0 := t.bandLimitedTbls[index].Hermite(x)
	v1 := t.bandLimitedTbls[index+1].Hermite(x)
	return v0*(1-offset) + v1*offset
}

func (t *Table) blTableIndexOffset(freq float64) (int, float64) {
	index := 0
	for i, f := range t.bandLimitedFreqs {
		if f > freq {
			break
		}
		index = i
	}
	offset := (freq - t.bandLimitedFreqs[index]) / (t.bandLimitedFreqs[index+1] - t.bandLimitedFreqs[index])
	return index, offset
}

func (t *Table) genBandLimited() {
	if t.bandLimitedTbls != nil {
		return
	}

	// minBandLimitedFreq * 2^((x - 1)*bandLimitedSemitoneRange/12) = maxBandLimitedFreq
	// x = 1 + 12 * log2(maxBandLimitedFreq/minBandLimitedFreq) / bandLimitedSemitoneRange
	numTables := int(1 + 12*math.Log2(maxBandLimitedFreq/minBandLimitedFreq)/bandLimitedSemitoneRange)
	t.bandLimitedTbls = make([]atomicTable, numTables)
	t.bandLimitedFreqs = make([]float64, numTables)
	for i := 0; i < numTables; i++ {
		freq := minBandLimitedFreq * math.Pow(2, float64(i)*bandLimitedSemitoneRange/12)
		if i == 0 && freq != minBandLimitedFreq {
			panic("wavtabs: internal error: first band-limited table has incorrect frequency")
		}
		if i == numTables-1 && freq != maxBandLimitedFreq {
			panic("wavtabs: internal error: last band-limited table has incorrect frequency")
		}
		tbl := t.bandLimited(freq, maxAudibleFrequency)
		t.bandLimitedTbls[i] = append(tbl, tbl[0])
		t.bandLimitedFreqs[i] = freq
	}
}

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
