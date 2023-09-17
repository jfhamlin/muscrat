package wavtabs

import (
	"sync"

	"gonum.org/v1/gonum/dsp/fourier"
)

var (
	ffts     = make(map[int]*fourier.FFT)
	fftsLock sync.Mutex
)

func (t *Table) fft() {
	if t.bins != nil {
		return
	}

	fftsLock.Lock()
	fft := ffts[t.Resolution()]
	if fft == nil {
		fft = fourier.NewFFT(t.Resolution())
		ffts[t.Resolution()] = fft
	}
	fftsLock.Unlock()

	t.bins = fft.Coefficients(nil, t.tbl[:t.Resolution()])
}
