package mod

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

type envOptions struct {
	interp      string
	releaseNode int
}

type EnvOption func(*envOptions)

func WithReleaseNode(node int) EnvOption {
	return func(o *envOptions) {
		o.releaseNode = node
	}
}

func WithInterpolation(interp string) EnvOption {
	return func(o *envOptions) {
		o.interp = interp
	}
}

func NewEnvelope(opts ...EnvOption) ugen.SampleGenerator {
	// Behavior follows that of SuperCollider's Env/EnvGen
	// https://doc.sccode.org/Classes/Env.html

	type Interp int
	const (
		LinInterp Interp = iota
		ExpInterp
		HoldInterp
		SustainInterp
	)

	o := envOptions{
		interp:      "lin",
		releaseNode: -1,
	}
	for _, opt := range opts {
		opt(&o)
	}

	interp := LinInterp
	switch o.interp {
	case "lin":
		interp = LinInterp
	case "exp":
		interp = ExpInterp
	case "hold":
		interp = HoldInterp
	default:
		panic(fmt.Sprintf("unknown interpolation type: %s", o.interp))
	}
	inputInterp := interp

	lastGate := false

	level := 0.0
	stage := 0
	delta := 0.0
	counter := 0

	setupStage := func(cfg ugen.SampleConfig, levels, times [][]float64, idx int) {
		stageLevel := levels[stage][idx]
		stageTime := times[stage-1][idx]

		interp = inputInterp
		counter = int(stageTime * float64(cfg.SampleRateHz))
		switch interp {
		case LinInterp:
			delta = (stageLevel - level) / float64(counter)
		case ExpInterp:
			delta = math.Pow(stageLevel/level, 1/float64(counter))
		case HoldInterp:
			level = stageLevel
			delta = 0
		}
	}

	return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		res := make([]float64, n)

		levels := getSampleArrays(cfg.InputSamples, "level")
		times := getSampleArrays(cfg.InputSamples, "time")
		gate := cfg.InputSamples["trigger"]
		if len(gate) == 0 {
			gate = make([]float64, n)
		}
		if len(levels) == 0 {
			return res
		}
		if len(levels) == 1 {
			for i := 0; i < n; i++ {
				res[i] = levels[0][i]
			}
			return res
		}

		for i := 0; i < n; i++ {
			if stage == 0 {
				level = levels[0][i]
			}

			if gate[i] > 0 && !lastGate {
				stage = 1
				setupStage(cfg, levels, times, i)
			}
			lastGate = gate[i] > 0

			if stage == 0 {
				res[i] = level
				continue
			}

			switch interp {
			case LinInterp:
				level += delta
			case ExpInterp:
				level *= delta
			case HoldInterp:
				// do nothing
			case SustainInterp:
				// do nothing
			}

			counter--
			if counter <= 0 {
				if stage < len(levels) {
					level = levels[stage][i]
				}
				if lastGate && stage == o.releaseNode {
					interp = SustainInterp
				} else {
					if stage+1 < len(levels) {
						stage++
						setupStage(cfg, levels, times, i)
					} else {
						interp = SustainInterp
					}
				}
			}
			res[i] = level
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
