package ugen

func NextPowerOf2(x int) int {
	p := 1
	for p < x {
		p <<= 1
	}
	return p
}

func LinInterp(t, x0, x1 float64) float64 {
	return x0 + t*(x1-x0)
}

func CubInterp(t, x0, x1, x2, x3 float64) float64 {
	c0 := x1
	c1 := 0.5 * (x2 - x0)
	c2 := x0 - 2.5*x1 + 2*x2 - 0.5*x3
	c3 := 0.5*(x3-x0) + 1.5*(x1-x2)

	return ((c3*t+c2)*t+c1)*t + c0
}
