package wavtabs

import "math"

// Tri returns a triangle wave table.
func Tri(resolution int) Table {
	tbl := make([]float64, resolution)
	for i := range tbl {
		tbl[i] = 1 - 2*math.Abs(2*float64(i)/float64(resolution)-1)
	}
	return tbl
}
