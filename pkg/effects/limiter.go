package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewLimiter(dur float64) ugen.UGen {
	// Port of supercollider's Limiter UGen
	// https://doc.sccode.org/Classes/Limiter.html

	// dur, aka lookAheadTime, is the buffer delay time. Shorter times
	// will produce smaller delays and quicker transient response times,
	// but may introduce amplitude modulation artifacts.

	bufsize := int(math.Ceil(dur * 44100))
	if bufsize < 1 {
		bufsize = 1
	}

	table := make([]float64, 3*bufsize)

	flips := 0
	pos := 0
	slope := 0.0
	level := 1.0
	prevMaxVal := 0.0
	curMaxVal := 0.0
	slopeFactor := 1.0 / float64(bufsize)

	xinBufFull := table[:bufsize]
	xmidBufFull := table[bufsize : 2*bufsize]
	xoutBufFull := table[2*bufsize : 3*bufsize]

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		// input signal
		in := cfg.InputSamples["in"]
		// The peak output amplitude level to which to normalize the input.
		amps := cfg.InputSamples["amp"]

		amp := amps[0]

		var val float64

		bufRemain := int(bufsize - pos)

		remain := len(out)
		for remain > 0 {
			offset := len(out) - remain

			nsmps := minInt(remain, bufRemain)
			xinBuf := xinBufFull[pos:]
			xoutBuf := xoutBufFull[pos:]
			if flips >= 2 {
				for i := 0; i < nsmps; i++ {
					val = in[offset+i]
					xinBuf[i] = val
					out[offset+i] = level * xoutBuf[i]
					level += slope
					val = math.Abs(val)
					if val > curMaxVal {
						curMaxVal = val
					}
				}
			} else {
				for i := 0; i < nsmps; i++ {
					val = in[offset+i]
					xinBuf[i] = val
					out[offset+i] = 0
					level += slope
					val = math.Abs(val)
					if val > curMaxVal {
						curMaxVal = val
					}
				}
			}
			pos += nsmps
			if pos >= bufsize {
				pos = 0
				bufRemain = bufsize

				maxVal2 := math.Max(prevMaxVal, curMaxVal)
				prevMaxVal = curMaxVal
				curMaxVal = 0.0

				var nextLevel float64
				if maxVal2 > amp {
					nextLevel = amp / maxVal2
				} else {
					nextLevel = 1.0
				}

				slope = (nextLevel - level) * slopeFactor

				temp := xoutBufFull
				xoutBufFull = xmidBufFull
				xmidBufFull = xinBufFull
				xinBufFull = temp

				flips++
			}
			remain -= nsmps
		}
	})
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
