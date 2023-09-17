package wavtabs

// Pulse returns a pulse wave table.
// TODO: pulse is not well-implemented by a wavetable. Drop it?
func Pulse(resolution int) *Table {
	return memoize("pulse", resolution, func() *Table {
		if resolution < 2 {
			panic("wavtabs: resolution must be at least 2")
		}
		tbl := make([]float64, resolution)
		for i := 0; i < resolution; i++ {
			tbl[i] = 1
		}
		return NewWithWrap(tbl, -1)
	})
}
