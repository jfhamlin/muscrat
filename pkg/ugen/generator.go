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

	// UGen is an abstract interface for generating samples.
	// Implementations should fill the output slice with samples. They
	// should *not* retain a reference to the output slice, as it may be
	// reused.
	UGen interface {
		Gen(ctx context.Context, cfg SampleConfig, out []float64)
	}

	// UGenFunc is a function that implements UGen.
	UGenFunc func(context.Context, SampleConfig, []float64)

	SimpleUGenFunc func(SampleConfig, []float64) []float64
)

func (gs UGenFunc) Gen(ctx context.Context, cfg SampleConfig, out []float64) {
	gs(ctx, cfg, out)
}

func (gs SimpleUGenFunc) Gen(ctx context.Context, cfg SampleConfig, out []float64) {
	gs(cfg, out)
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
