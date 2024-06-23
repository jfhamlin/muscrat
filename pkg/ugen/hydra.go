package ugen

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type (
	// NewHydra creates a new Hydra unit generator. It's a false ugen that
	// returns zero samples. It's used to update the Hydra synthesizer
	hydra struct {
		graph any
	}
)

func NewHydra(graph any) UGen {
	return &hydra{
		graph: graph,
	}
}

func (h *hydra) Start(ctx context.Context) error {
	go runtime.EventsEmit(ctx, "hydra.graph", h.graph)
	return nil
}

func (h *hydra) Gen(ctx context.Context, cfg SampleConfig, out []float64) {
	// collect the latest value from all inputs and send as
	// hydra.mappings
	mappings := make(map[string]float64)
	idx := len(out) - 1
	for k, v := range cfg.InputSamples {
		mappings[k] = v[idx]
	}
	go runtime.EventsEmit(ctx, "hydra.mapping", mappings)
	return
}
