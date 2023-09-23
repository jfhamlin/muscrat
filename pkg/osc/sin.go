package osc

import (
	"math"

	"github.com/jfhamlin/muscrat/pkg/interp"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

const (
	sineTableSize = 2048
)

var (
	// sineTable is a lookup table for sine values. It includes 2048
	// samples, covering [0, 1), with samples duplicated at the start
	// and end to allow for wrapping without modulo.
	sineTable = [1 + sineTableSize + 2]float64{}
)

func init() {
	for i := 0; i < sineTableSize; i++ {
		sineTable[i+1] = math.Sin(float64(i) / sineTableSize * 2 * math.Pi)
	}
	sineTable[0] = sineTable[sineTableSize]
	sineTable[sineTableSize+1] = sineTable[1]
	sineTable[sineTableSize+2] = sineTable[2]
}

func NewSine(opts ...ugen.Option) ugen.UGen {
	return New(SamplerFunc(sampleSine), opts...)
}

func sampleSine(phase, dPhase, dutyCycle float64) float64 {
	phase /= dutyCycle
	phase = math.Max(0, math.Min(1, phase))

	x := phase - math.Floor(phase)
	x = x * float64(sineTableSize)
	i := int(x)
	off := x - float64(i)
	val0 := sineTable[i]
	val1 := sineTable[i+1]
	val2 := sineTable[i+2]
	val3 := sineTable[i+3]
	return interp.Hermite(off, val0, val1, val2, val3)
}
