package wavtabs

// Saw is a sawtooth wave table.
func Saw(resolution int) Table {
	table := make([]float64, resolution)
	for i := 0; i < resolution; i++ {
		table[i] = float64(i) / float64(resolution)
	}
	return table
}
