package ugen

import (
	"math"
	"strconv"
	"strings"
)

// CollectIndexedInputs collects all inputs of the form "$<int>".
func CollectIndexedInputs(cfg SampleConfig) [][]float64 {
	res := make([][]float64, 0, len(cfg.InputSamples))
	for k, v := range cfg.InputSamples {
		if !strings.HasPrefix(k, "$") {
			continue
		}
		idx, err := strconv.Atoi(k[1:])
		if err != nil {
			continue
		}
		if idx >= len(res) {
			for i := len(res); i <= idx; i++ {
				res = append(res, nil)
			}
		}
		res[idx] = v
	}
	return res
}

func ZapGremlins(x float64) float64 {
	absx := math.Abs(x)
	//     // very small numbers fail the first test, eliminating denormalized numbers
	//    (zero also fails the first test, but that is OK since it returns zero.)
	// very large numbers fail the second test, eliminating infinities
	// Not-a-Numbers fail both tests and are eliminated.
	// return (absx > (float32)1e-15 && absx < (float32)1e15) ? x : (float32)0.;

	if absx > 1e-15 && absx < 1e15 {
		return x
	}
	return 0
}
