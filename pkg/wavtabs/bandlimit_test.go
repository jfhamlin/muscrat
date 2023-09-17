package wavtabs

import (
	"fmt"
	"math"
	"math/cmplx"
	"testing"

	"gonum.org/v1/gonum/dsp/fourier"
)

func TestBandLimit(t *testing.T) {
	return

	const sz = 8
	sig := make([]float64, sz)
	for i := range sig {
		sig[i] = math.Sin(2 * math.Pi * float64(i) / sz)
		fmt.Printf("%v\n", sig[i])
	}
	fmt.Println("=========")

	fft := fourier.NewFFT(sz)
	res := make([]complex128, sz/2+1)
	fft.Coefficients(res, sig)
	for i := range res {
		fmt.Printf("%v\n", cmplx.Abs(res[i]))
		if i != 1 {
			res[i] = 0
		}
	}

	fmt.Println("=========")
	fft.Sequence(sig, res)
	for i := range sig {
		fmt.Printf("%v\n", sig[i]/float64(sz))
	}
}
