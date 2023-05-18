package ugen

import "context"

// SampleConfig is a configuration for a sample generator.
type SampleConfig struct {
	// The sample rate of the output stream.
	SampleRateHz int

	// Input samples that can be used to generate the output samples.
	InputSamples map[string][]float64
}

// SampleGenerator is an abstract interface for generating samples.
type SampleGenerator interface {
	GenerateSamples(ctx context.Context, cfg SampleConfig, n int) []float64
}

// SampleGeneratorFunc is a function that implements SampleGenerator.
type SampleGeneratorFunc func(context.Context, SampleConfig, int) []float64

func (gs SampleGeneratorFunc) GenerateSamples(ctx context.Context, cfg SampleConfig, n int) []float64 {
	return gs(ctx, cfg, n)
}

// Starter is an interface for starting a sample generator. If a
// sample generator implements this interface, it will be started when
// a graph is run.
type Starter interface {
	Start() error
}

// Stopper is an interface for stopping a sample generator. If a
// sample generator implements this interface, it will be stopped when
// a graph is stopped.
type Stopper interface {
	Stop() error
}
