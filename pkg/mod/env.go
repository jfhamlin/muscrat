package mod

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/jfhamlin/muscrat/pkg/bufferpool"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

const (
	shapeLin     = "lin"
	shapeCurve   = "curve"
	shapeExp     = "exp"
	shapeHold    = "hold"
	shapeSustain = "sustain"
)

type (
	envOptions struct {
		// curve describes the interpolation of each stage its length should
		// be less than or equal to the number of stages. if less, curve
		// values will wrap around for any additional stages.
		// values can be strings or floats
		curve       []any
		releaseNode int
	}
)

type EnvOption func(*envOptions)

func WithReleaseNode(node int) EnvOption {
	return func(o *envOptions) {
		o.releaseNode = node
	}
}

func WithCurve(curve []any) EnvOption {
	return func(o *envOptions) {
		o.curve = curve
	}
}

func NewEnvelope(opts ...EnvOption) ugen.UGen {
	// Behavior more or less follows that of SuperCollider's Env/EnvGen
	// https://doc.sccode.org/Classes/Env.html

	o := envOptions{
		curve:       []any{"lin"},
		releaseNode: -1,
	}
	for _, opt := range opts {
		opt(&o)
	}

	// validate the curve values
	for _, c := range o.curve {
		switch c := c.(type) {
		case string:
			switch c {
			case shapeLin, shapeExp, shapeHold:
			default:
				panic(fmt.Sprintf("invalid curve value: %s", c))
			}
		case float64:
		default:
			panic(fmt.Sprintf("invalid curve value: %v (%T)", c, c))
		}
	}

	shape := shapeLin

	lastGate := false

	level := 0.0
	stage := 0
	delta := 0.0
	counter := 0

	// for curve interpolation
	var a2, b1 float64

	setupStage := func(cfg ugen.SampleConfig, levels, times [][]float64, idx int) {
		stageLevel := levels[stage][idx]
		stageTime := times[stage-1][idx]
		stageShape := o.curve[(stage-1)%len(o.curve)]

		var curve float64
		switch s := stageShape.(type) {
		case string:
			shape = s
		case float64:
			shape = shapeCurve
			curve = s
		default:
			panic(fmt.Sprintf("invalid curve value: %v (%T)", stageShape, stageShape))
		}

		counter = int(stageTime * float64(cfg.SampleRateHz))
		switch shape {
		case shapeLin:
			delta = (stageLevel - level) / float64(counter)
		case shapeCurve:
			if math.Abs(curve) < 0.001 {
				shape = shapeLin
				delta = (stageLevel - level) / float64(counter)
			} else {
				a1 := (stageLevel - level) / (1 - math.Exp(curve))
				a2 = level + a1
				b1 = a1
				delta = math.Exp(curve / float64(counter))
			}
		case shapeExp:
			delta = math.Pow(stageLevel/level, 1/float64(counter))
		case shapeHold:
			level = levels[stage-1][idx]
			delta = 0
		}
	}

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		n := len(out)

		levels := getSampleArrays(cfg.InputSamples, "level")
		times := getSampleArrays(cfg.InputSamples, "time")
		gate := cfg.InputSamples["trigger"]
		if len(gate) == 0 {
			buf := bufferpool.Get(n)
			gate = *buf
			defer bufferpool.Put(buf)
		}
		if len(levels) == 0 {
			return
		}
		if len(levels) == 1 {
			for i := 0; i < n; i++ {
				out[i] = levels[0][i]
			}
			return
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
				out[i] = level
				continue
			}

			switch shape {
			case shapeLin:
				level += delta
			case shapeCurve:
				b1 *= delta
				level = a2 - b1
			case shapeExp:
				level *= delta
			case shapeHold:
				// do nothing
			case shapeSustain:
				// do nothing
			}

			counter--
			if counter <= 0 {
				if stage < len(levels) {
					level = levels[stage][i]
				}
				if lastGate && stage == o.releaseNode {
					shape = shapeSustain
				} else {
					if stage+1 < len(levels) {
						stage++
						setupStage(cfg, levels, times, i)
					} else {
						shape = shapeSustain
						if stage < len(levels) {
							level = levels[stage][i]
						}
					}
				}
			}
			out[i] = level
		}
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
