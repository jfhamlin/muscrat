package wavtabs

import "math"

// Sin returns a sin wave table.
func Sin(resolution int) Table {
	sin := make([]float64, resolution)
	for i := range sin {
		sin[i] = math.Sin(2 * math.Pi * float64(i) / float64(resolution))
	}
	return sin
}
