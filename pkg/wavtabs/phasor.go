package wavtabs

// Phasor is a sawtooth wave table in the range [0, 1].
// TODO: phasor is not well-implemented by a wavetable. Drop it?
func Phasor(resolution int) *Table {
	return memoize("phasor", resolution, func() *Table {
		table := make([]float64, resolution)
		for i := 0; i < len(table); i++ {
			table[i] = float64(i) / float64(resolution-1)
		}
		return New(table)
	})
}
