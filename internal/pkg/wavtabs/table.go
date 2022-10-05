package wavtabs

const (
	DefaultResolution = 1024
)

// Table is a wavetable.
type Table []float64

// Lerp linearly interpolates between two discrete values in a
// wavetable.
func (t Table) Lerp(x float64) float64 {
	x = x * float64(len(t))
	i := int(x)
	f := x - float64(i)
	return t[i%len(t)]*(1-f) + t[(i+1)%len(t)]*f
}
