package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/freeverb-go"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewFreeverb(revmod *freeverb.RevModel) ugen.SampleGenerator {
	return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		input32 := make([]float32, n)
		for i := 0; i < n; i++ {
			input32[i] = float32(cfg.InputSamples["$0"][i])
		}
		if roomSizes, ok := cfg.InputSamples["room-size"]; ok {
			rs := roomSizes[0]
			roomSize := float32(math.Max(0, math.Min(1, rs)))
			if math.Abs(float64(roomSize-revmod.GetRoomSize())) > 0.01 {
				revmod.SetRoomSize(roomSize)
			}
		}
		if drys, ok := cfg.InputSamples["dry"]; ok {
			d := drys[0]
			dry := float32(math.Max(0, math.Min(1, d)))
			if math.Abs(float64(dry-revmod.GetDry())) > 0.01 {
				revmod.SetDry(dry)
			}
		}

		outputLeft := make([]float32, n)
		outputRight := make([]float32, n)
		revmod.ProcessReplace(input32, input32, outputLeft, outputRight, n, 1)
		output := make([]float64, n)
		for i := 0; i < n; i++ {
			output[i] = 0.5 * (float64(outputLeft[i]) + float64(outputRight[i]))
		}

		return output
	})
}
