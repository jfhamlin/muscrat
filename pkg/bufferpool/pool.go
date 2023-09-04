package bufferpool

import "sync"

var (
	pool = sync.Pool{}
)

func Get(sz int) []float64 {
	v := pool.Get()
	if v == nil {
		return make([]float64, sz, sz)
	}
	vslc := v.([]float64)
	if len(vslc) < sz {
		return make([]float64, sz, sz)
	}
	clear(vslc)
	return vslc[:sz]
}

func Put(buf []float64) {
	pool.Put(buf)
}
