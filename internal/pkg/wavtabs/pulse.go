package wavtabs

// Pulse returns a pulse wave table.
func Pulse(resolution int) Table {
	res := make([]float64, resolution+1)
	for i := 0; i < resolution; i++ {
		res[i] = 1
	}
	res[resolution] = -1
	return res
}
