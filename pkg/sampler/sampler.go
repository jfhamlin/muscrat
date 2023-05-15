package sampler

import (
	"context"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewSampler(buf []float64, loop bool) ugen.SampleGenerator {
	if len(buf) == 0 {
		return ugen.NewConstant(0)
	}

	// TODO: sampler should accept a trigger
	// TODO: sampler should accept arg for sample rate

	sampleLen := len(buf)
	index := 0
	stopped := false
	return ugen.SampleGeneratorFunc(func(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := 0; i < n; i++ {
			if stopped {
				res[i] = 0
				continue
			}

			if index >= sampleLen {
				res[i] = 0
			} else {
				res[i] = buf[index]
			}

			index++
			if index >= sampleLen {
				index = 0
				if !loop {
					stopped = true
				}
			}
		}
		return res
	})
}
