package wavtabs

import (
	"context"
	"testing"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

// BenchmarkGenerator-8   	  208602	      5770 ns/op // Before bandlimiting

// BenchmarkGenerator-8   	   33162	     37659 ns/op
func BenchmarkGenerator(b *testing.B) {
	g := Generator(Sin(2048))

	const numSamples = 1024
	w := make([]float64, numSamples)
	for i := range w {
		w[i] = 440
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
