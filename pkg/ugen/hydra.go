package ugen

import (
	"context"

	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

type (
	// NewHydra creates a new Hydra unit generator. It's a false ugen that
	// returns zero samples. It's used to update the Hydra synthesizer
	hydra struct {
		Expr any      `json:"expr"`
		Vars []string `json:"vars"`
	}
)

func NewHydra(expr any, vars []string) UGen {
	return &hydra{
		Expr: expr,
		Vars: vars,
	}
}

func (h *hydra) Start(ctx context.Context) error {
	go pubsub.Publish("hydra.expr", h)
	return nil
}

func (h *hydra) Gen(ctx context.Context, cfg SampleConfig, out []float64) {
	// collect the latest value from all inputs and send as
	// hydra.mappings
	mappings := make(map[string]float64)
	idx := len(out) - 1
	for k, v := range cfg.InputSamples {
		val := v[idx]
		// if NaN, zero it
		if val != val {
			val = 0
		}
		mappings[k] = val
	}
	go pubsub.Publish("hydra.mapping", mappings)
	return
}
