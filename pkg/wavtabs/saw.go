package wavtabs

// Saw is a sawtooth wave table.
func Saw(resolution int) *Table {
	table := make([]float64, resolution)
	for i := range table {
		table[i] = 2*float64(i)/float64(resolution-1) - 1
	}
	return New(table)
}
