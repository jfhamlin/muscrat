package slice

func FindIndexOfRisingEdge(s []float64, start int, lastVal float64) int {
	for i := start; i < len(s); i++ {
		if lastVal <= 0 && s[i] > 0 {
			return i
		}
	}
	return len(s)
}
