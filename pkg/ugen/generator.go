package ugen

import "context"

type (
	// SampleConfig is a configuration for a sample generator.
	SampleConfig struct {
		// The sample rate of the output stream.
		SampleRateHz int

		// Input samples that can be used to generate the output samples.
		InputSamples map[string][]float64
	}

	// SampleGenerator is an abstract interface for generating samples.
	SampleGenerator interface {
		GenerateSamples(ctx context.Context, cfg SampleConfig, n int) []float64
	}

	// SampleGeneratorFunc is a function that implements SampleGenerator.
	SampleGeneratorFunc func(context.Context, SampleConfig, int) []float64

	SimpleSampleGeneratorFunc func(SampleConfig, int) []float64

	UGen     = SampleGenerator
	UGenFunc = SampleGeneratorFunc
)

func (gs SampleGeneratorFunc) GenerateSamples(ctx context.Context, cfg SampleConfig, n int) []float64 {
	return gs(ctx, cfg, n)
}

func (gs SimpleSampleGeneratorFunc) GenerateSamples(ctx context.Context, cfg SampleConfig, n int) []float64 {
	return gs(cfg, n)
}

// Starter is an interface for starting a sample generator. If a
// sample generator implements this interface, it will be started when
// a graph is run.
type Starter interface {
	Start(ctx context.Context) error
}

// Stopper is an interface for stopping a sample generator. If a
// sample generator implements this interface, it will be stopped when
// a graph is stopped.
type Stopper interface {
	Stop(ctx context.Context) error
}
