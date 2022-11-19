package vector

import "testing"

func TestTransient(t *testing.T) {
	trans := newTransient(&vector{})

	const n = 10000

	for i := 0; i < n; i++ {
		trans.conj(i)
	}
	vec := trans.persistent()
	for i := 0; i < n; i++ {
		val, ok := vec.Index(i)
		if !ok {
			t.Errorf("Index %d not found", i)
		}
		if val != i {
			t.Errorf("Index %d has value %d, expected %d", i, val, i)
		}
	}
}
