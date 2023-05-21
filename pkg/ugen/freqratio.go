package ugen

import (
	"context"
	"fmt"
	"math"
)

func NewFreqRatio(typ string) SampleGenerator {
	divisor := 12.0
	switch typ {
	case "semitones":
		divisor = 12.0
	default:
		panic(fmt.Errorf("unknown freq ratio type: %s", typ))
	}
	return SampleGeneratorFunc(func(ctx context.Context, cfg SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := range res {
			res[i] = 1
			for _, s := range cfg.InputSamples {
				res[i] *= math.Pow(2, s[i]/divisor)
			}
		}
		return res
	})
}
