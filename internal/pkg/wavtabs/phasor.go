package wavtabs

// Phasor is a sawtooth wave table in the range [0, 1).
func Phasor(resolution int) Table {
	table := make([]float64, resolution)
	for i := 0; i < resolution; i++ {
		table[i] = float64(i) / float64(resolution)
	}
	return table
}
