package ugen

import (
	"context"
	"math"
)

func NewMovingAverage(maxDurSecs float64) UGen {
	maxDurSecs = math.Max(0.01, maxDurSecs)

	var (
		buf     []float64
		sum     float64
		curSize int

		lastDurS      float64
		windowSize    int
		maxWindowSize int

		head, tail int
	)

	return UGenFunc(func(ctx context.Context, cfg SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		dur := cfg.InputSamples["dur"]

		// index the last elements to lift the bounds check
		_ = in[len(out)-1]
		_ = dur[len(out)-1]

		if len(buf) == 0 {
			maxWindowSize = int(maxDurSecs * float64(cfg.SampleRateHz))
			if maxWindowSize < 1 {
				maxWindowSize = 1
			}
			buf = make([]float64, maxWindowSize)
			// fill the buffer with the first value
			for i := range buf {
				buf[i] = in[0]
			}
			lastDurS = math.Min(dur[0], maxDurSecs)
			windowSize = int(lastDurS * float64(cfg.SampleRateHz))
			if windowSize < 1 {
				windowSize = 1
			}
		}

		for i := range out {
			// update the window size if the duration has changed
			newDur := math.Min(dur[i], maxDurSecs)
			if newDur != lastDurS {
				lastDurS = newDur
				windowSize = int(lastDurS * float64(cfg.SampleRateHz))
				if windowSize < 1 {
					windowSize = 1
				}
			}

			for curSize >= windowSize {
				sum -= buf[tail]
				tail++
				if tail == maxWindowSize {
					tail = 0
				}
				curSize--
			}

			sum += in[i]
			buf[head] = in[i]
			curSize++

			head++
			if head == maxWindowSize {
				head = 0
			}

			out[i] = sum / float64(curSize)
		}
	})
}
