package interp

// Hermite performs Hermite interpolation on the given values, where
// off is the offset from val1 towards val2. off is assumed to be in
// the range [0, 1).
func Hermite(off, val0, val1, val2, val3 float64) float64 {
	slope0 := (val2 - val0) * 0.5
	slope1 := (val3 - val1) * 0.5
	v := val1 - val2
	w := slope0 + v
	a := w + v + slope1
	bNeg := w + a
	stage1 := a*off - bNeg
	stage2 := stage1*off + slope0
	return stage2*off + val1
}
