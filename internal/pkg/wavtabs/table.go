package wavtabs

import "math"

const (
	DefaultResolution = 1024
)

// Table is a wavetable.
type Table []float64

// Lerp linearly interpolates between two discrete values in a
// wavetable. The function described by the table is assumed to be
// periodic, with the contents describing a single period of the
// waveform covering [0, 1). Values of x outside this range are valid,
// and will be wrapped to the appropriate position in the table.
func (t Table) Lerp(x float64) float64 {
	x -= math.Floor(x)
	x = x * float64(len(t))
	i := int(x)
	f := x - float64(i)
	return t[i%len(t)]*(1-f) + t[(i+1)%len(t)]*f
}
