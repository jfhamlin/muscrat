package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/freeverb-go"
	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewFreeverb(revmod *freeverb.RevModel) ugen.UGen {
	input32 := make([]float32, conf.BufferSize)
	outputLeft := make([]float32, conf.BufferSize)
	outputRight := make([]float32, conf.BufferSize)

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		n := len(out)
		if len(input32) != n {
			input32 = make([]float32, n)
			outputLeft = make([]float32, n)
			outputRight = make([]float32, n)
		}

		for i := range input32 {
			input32[i] = float32(cfg.InputSamples["in"][i])
		}
		if roomSizes, ok := cfg.InputSamples["room-size"]; ok {
			rs := roomSizes[0]
			roomSize := float32(math.Max(0, math.Min(1, rs)))
			if math.Abs(float64(roomSize-revmod.GetRoomSize())) > 0.01 {
				revmod.SetRoomSize(roomSize)
			}
		}
		if damps, ok := cfg.InputSamples["damp"]; ok {
			dmp := damps[0]
			damp := float32(math.Max(0, math.Min(1, dmp)))
			if math.Abs(float64(damp-revmod.GetDamp())) > 0.01 {
				revmod.SetDamp(damp)
			}
		}
		if mixes, ok := cfg.InputSamples["mix"]; ok {
			mix := math.Max(0, math.Min(1, mixes[0])) // dry/wet
			dry := float32(1 - mix)
			wet := float32(mix)
			if math.Abs(float64(dry-revmod.GetDry())) > 0.01 {
				revmod.SetDry(dry)
				revmod.SetWet(wet)
			}
		}

		// TODO: multi-channel output
		revmod.ProcessReplace(input32, input32, outputLeft, outputRight, n, 1)
		for i := range out {
			out[i] = 0.5 * (float64(outputLeft[i]) + float64(outputRight[i]))
		}
	})
}
