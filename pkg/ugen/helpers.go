package ugen

import (
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
