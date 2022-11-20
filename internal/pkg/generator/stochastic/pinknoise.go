package stochastic

import (
	"context"
	"math/bits"
	"math/rand"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
)

const (
	numVMCOctaves = 16 // must be a power of 2
)

// PinkNoise is a stochastic generator that produces pink noise (1/f
// noise) using the Voss-McCartney algorithm.
type PinkNoise struct {
	counter uint
	total   float64
	// dice[k] is the previous value of the kth octave
	dice [numVMCOctaves]float64
	rand *rand.Rand
}

// NewPinkNoise returns a new PinkNoise stochastic generator.
func NewPinkNoise(opts ...Option) *PinkNoise {
	o := options{
		rand: rand.New(rand.NewSource(0)),
	}
	for _, opt := range opts {
		opt(&o)
	}

	dice := [numVMCOctaves]float64{}
	for i := range dice {
		dice[i] = pinkNoiseRandom(o.rand)
	}
	return &PinkNoise{
		rand: o.rand,
		dice: dice,
	}
}

func (pn *PinkNoise) GenerateSamples(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
	samples := make([]float64, n)
	for i := 0; i < n; i++ {
		samples[i] = pn.generateSample()
	}
	return samples
}

// generateSample generates a single sample of pink noise.
func (pn *PinkNoise) generateSample() float64 {
	k := bits.TrailingZeros(pn.counter)
	k = k & (numVMCOctaves - 1)

	pn.counter++

	// get previous value of this octave
	prevrand := pn.dice[k]

	// generate a new random value
	newrand := pinkNoiseRandom(pn.rand)

	// store new value
	pn.dice[k] = newrand

	// update total
	pn.total += (newrand - prevrand)

	// generate a new random value for the top octave
	newrand = pinkNoiseRandom(pn.rand)

	return (pn.total + newrand)
}

func pinkNoiseRandom(r *rand.Rand) float64 {
	return 0.5*r.Float64() - 0.25
}
