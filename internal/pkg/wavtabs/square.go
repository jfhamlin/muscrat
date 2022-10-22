package wavtabs

// Square returns a square wave table.
func Square(resolution int, dutyCycle float64) Table {
	table := make(Table, resolution+1)
	for i := 0; i < resolution; i++ {
		if float64(i) < float64(resolution)*dutyCycle {
			table[i] = 1
		} else {
			table[i] = -1
		}
	}
	table[resolution] = table[0]
	return table
}
