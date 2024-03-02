package ugen

import (
	"context"
	"fmt"
	"math"
)

func NewFreqRatio(typ string) UGen {
	base := 2.0
	divisor := 12.0
	switch typ {
	case "octaves":
		divisor = 1.0
	case "semitones":
		divisor = 12.0
	case "cents":
		divisor = 1200.0
	case "decibels":
		base = 10.0
		divisor = 20.0
	default:
		panic(fmt.Errorf("unknown freq ratio type: %s", typ))
	}
	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		for i := range out {
			out[i] = 1
		}
		for _, s := range cfg.InputSamples {
			// index the last element of s to lift the bounds check
			_ = s[len(out)-1]
			for i := range out {
				out[i] *= math.Pow(base, s[i]/divisor)
			}
		}
	})
}
