package mod

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewEnvelope(interpolation string) ugen.SampleGenerator {
	// Behavior follows that of SuperCollider's Env/EnvGen
	// https://doc.sccode.org/Classes/Env.html

	triggered := false
	triggerTime := 0.0
	lastGate := false
	return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		res := make([]float64, n)

		levels := getSampleArrays(cfg.InputSamples, "level")
		times := getSampleArrays(cfg.InputSamples, "time")
		gate := cfg.InputSamples["trigger"]
		if len(gate) == 0 {
			gate = make([]float64, n)
		}

		for i := 0; i < n; i++ {
			var envDur float64
			for _, t := range times {
				envDur += t[i]
			}

			if (!triggered || triggerTime > envDur) && gate[i] > 0 && !lastGate {
				triggered = true
				triggerTime = 0
			}

			lastGate = gate[i] > 0

			if !triggered {
				res[i] = levels[0][i]
				continue
			}
			if triggerTime > envDur {
				res[i] = levels[len(levels)-1][i]
				continue
			}

			// interpolate between the two levels adjacent to the current
			// time.
			var timeSum float64
			for j, t := range times {
				timeSum += t[i]
				if timeSum >= triggerTime {
					level1 := levels[j][i]
					level2 := levels[j+1][i]
					// interpolate between levels[j] and levels[j+1]
					// at time triggerTime
					switch interpolation {
					case "lin":
						res[i] = level1 + (level2-level1)*(triggerTime-(timeSum-t[i]))/t[i]
					case "exp":
						res[i] = level1 * math.Pow(level2/level1, (triggerTime-(timeSum-t[i]))/t[i])
					}
					break
				}
			}
			triggerTime += 1 / float64(cfg.SampleRateHz)
		}
		return res
	})
}

func getSampleArrays(inputs map[string][]float64, name string) [][]float64 {
	var numSamples int

	prefix := name + "$"
	var res [][]float64
	for key, in := range inputs {
		if strings.HasPrefix(key, prefix) {
			idx, err := strconv.Atoi(key[len(prefix):])
			if err != nil {
				continue
			}
			for idx >= len(res) {
				res = append(res, nil)
			}
			res[idx] = in
			if numSamples == 0 {
				numSamples = len(in)
			}
		}
	}
	// fill in any missing inputs with arrays of zeros.
	for i, in := range res {
		if len(in) == 0 {
			res[i] = make([]float64, numSamples)
		}
	}
	return res
}
