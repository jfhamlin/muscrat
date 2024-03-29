package osc

import (
	"context"
	"testing"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

const (
	numTestSamples = 1024
)

func BenchmarkSine(b *testing.B) {
	benchmark(b, NewSine())
}

func BenchmarkSaw(b *testing.B) {
	benchmark(b, NewSaw())
}

func BenchmarkPulse(b *testing.B) {
	benchmark(b, NewPulse())
}

func BenchmarkTri(b *testing.B) {
	benchmark(b, NewTri())
}

func BenchmarkPhasor(b *testing.B) {
	benchmark(b, NewPhasor())
}

func benchmark(b *testing.B, osc ugen.UGen) {
	cfg := ugen.SampleConfig{
		SampleRateHz: 44100,
	}
	out := make([]float64, numTestSamples)
	for i := 0; i < b.N; i++ {
		osc.Gen(context.Background(), cfg, out)
	}
}
