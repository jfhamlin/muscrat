package ugen

import (
	"context"
	"fmt"
	"math"
)

func NewFreqRatio(typ string) SampleGenerator {
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
	return SampleGeneratorFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := range res {
			res[i] = 1
			for _, s := range cfg.InputSamples {
				res[i] *= math.Pow(base, s[i]/divisor)
			}
		}
		return res
	})
}
