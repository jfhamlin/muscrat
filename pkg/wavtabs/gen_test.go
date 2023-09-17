package wavtabs

import (
	"context"
	"math"
	"math/rand"
	"testing"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

// BenchmarkGenerator-8   	  208602	      5770 ns/op // Before bandlimiting

// BenchmarkGenerator-8   	   33162	     37659 ns/op // before bl table search optimization

// BenchmarkGenerator-8   	   42739	     25936 ns/op // before even more bl table search optimization

// BenchmarkGenerator-8   	   89032	     13406 ns/op
func BenchmarkGenerator(b *testing.B) {
	g := Generator(Sin(2048))

	const (
		numSamples  = 1024
		maxTestFreq = 20000.0
	)

	testCoefficient := math.Log2(maxTestFreq)
	w := make([]float64, numSamples)
	for i := range w {
		w[i] = math.Pow(2, rand.Float64()*testCoefficient)
	}

	out := make([]float64, numSamples)
	for i := 0; i < b.N; i++ {
		g.Gen(context.Background(), ugen.SampleConfig{
			SampleRateHz: 44100,
			InputSamples: map[string][]float64{
				"w": w,
			},
		}, out)
	}
}
