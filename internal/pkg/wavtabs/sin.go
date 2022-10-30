package wavtabs

import "math"

// Sin returns a sine wave table.
func Sin(resolution int) Table {
	sin := make([]float64, resolution+1)
	for i := range sin {
		sin[i] = math.Sin(2 * math.Pi * float64(i) / float64(resolution))
	}
	return sin
}
