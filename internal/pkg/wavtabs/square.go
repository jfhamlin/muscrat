package wavtabs

// Square returns a square wave table.
func Square(resolution int, dutyCycle float64) Table {
	table := make(Table, resolution)
	for i := range table {
		if float64(i) < float64(resolution)*dutyCycle {
			table[i] = 1
		} else {
			table[i] = -1
		}
	}
	return table
}
