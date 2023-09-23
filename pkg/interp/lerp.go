package interp

// Lerp performs linear interpolation between a and b, where t is the
// interpolation factor. t is assumed to be in the range [0, 1).
func Lerp(t, a, b float64) float64 {
	return a + t*(b-a)
}
