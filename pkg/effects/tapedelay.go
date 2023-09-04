package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewTapeDelay() ugen.UGen {
	// Simulate a tape delay by using a buffer of samples with a read
	// and write pointer. If the delay is changed, we simulate a
	// physical read/write head by maintaining a sample velocity for the
	// read head. The write head is always at the end of the buffer. The
	// read head can never move backwards, so if the delay is decreased,
	// the read head will accelerate, and if the delay is increased, the
	// read head will decelerate.
	var tape []float64
	var readHead float64
	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["$0"]
		delays := cfg.InputSamples["delay"]

		for i := range out {
			delaySeconds := delays[i]
			if delaySeconds < 0 {
				delaySeconds = 0
			}
			delaySamples := delaySeconds * float64(cfg.SampleRateHz)
			// handle the initialization case, where the tape hasn't been set up yet.
			if tape == nil {
				tape = make([]float64, int(delaySeconds*float64(cfg.SampleRateHz)))
			}
			actualDelaySamples := float64(len(tape)) - readHead

			tape = append(tape, in[i])

			if len(tape) == 1 {
				out[i] = tape[0]
			} else {
				// read the sample from the tape at the read head with linear interpolation
				// between the two adjacent samples.
				readHeadInt, readHeadFrac := math.Modf(readHead)
				out[i] = tape[int(readHeadInt)]*(1-readHeadFrac) + tape[int(readHeadInt)+1]*readHeadFrac
			}

			const maxStep = 2
			const minStep = 1 / maxStep

			// update the read head position with max and min bounds to prevent
			// the read head from moving backwards or infinitely forward.
			if delaySamples == 0 && actualDelaySamples > 0 {
				readHead += maxStep
			} else if actualDelaySamples > maxStep*delaySamples {
				readHead += maxStep
			} else if actualDelaySamples < minStep*delaySamples {
				readHead += minStep
			} else {
				vel := actualDelaySamples / delaySamples
				if math.IsNaN(vel) {
					readHead += maxStep
				} else {
					readHead += math.Max(minStep, math.Min(maxStep, vel))
				}
			}
			if readHead >= float64(len(tape)) {
				readHead = 0
				tape = tape[:0]
			}
			// drop samples that have already been read from the tape.
			if readHead > 1 {
				tape = tape[int(readHead):]
				readHead = readHead - math.Floor(readHead)
			}
		}
	})
}
