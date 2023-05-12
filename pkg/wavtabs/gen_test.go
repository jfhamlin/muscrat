package wavtabs

import (
	"context"
	"testing"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func BenchmarkGenerator(b *testing.B) {
	g := Generator(Sin(1024))

	const numSamples = 1024
	w := make([]float64, numSamples)
	for i := range w {
		w[i] = 440
	}

	for i := 0; i < b.N; i++ {
		g.GenerateSamples(context.Background(), ugen.SampleConfig{
			SampleRateHz: 44100,
			InputSamples: map[string][]float64{
				"w": w,
			},
		}, numSamples)
	}
}
