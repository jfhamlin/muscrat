package wavtabs

// Saw is a sawtooth wave table.
func Saw(resolution int) Table {
	table := make([]float64, resolution)
	for i := 0; i < resolution; i++ {
		table[i] = 2*float64(i)/float64(resolution) - 1
	}
	return table
}
