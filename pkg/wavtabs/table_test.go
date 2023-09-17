package wavtabs

import "testing"

// BenchmarkNearest-8     	1000000000	         0.8451 ns/op
func BenchmarkNearest(b *testing.B) {
	tbl := Sin(DefaultResolution)
	for i := 0; i < b.N; i++ {
		tbl.Nearest(float64(i) / float64(b.N))
	}
}

// BenchmarkLerp-8        	1000000000	         0.9042 ns/op
func BenchmarkLerp(b *testing.B) {
	tbl := Sin(DefaultResolution)
	for i := 0; i < b.N; i++ {
		tbl.Lerp(float64(i) / float64(b.N))
	}
}

// BenchmarkHermite-8     	405747754	         2.963 ns/op
func BenchmarkHermite(b *testing.B) {
	tbl := Sin(DefaultResolution)
	for i := 0; i < b.N; i++ {
		tbl.Hermite(float64(i) / float64(b.N))
	}
}
