package bufferpool

import (
	"fmt"
	"math/bits"
	"sync"
)

const (
	MinSize = 1 << 7  // 128
	MaxSize = 1 << 11 // 2048
)

var (
	pools []*sync.Pool
)

func init() {
	for i := MinSize; i <= MaxSize; i *= 2 {
		pools = append(pools, &sync.Pool{})
	}
}

func isPowerOfTwo(n int) bool {
	return n&(n-1) == 0
}

func Get(sz int) *[]float64 {
	if sz < MinSize || sz > MaxSize || !isPowerOfTwo(sz) {
		panic(fmt.Sprintf("invalid size: %d", sz))
	}
	pool := pools[bits.TrailingZeros(uint(sz)>>7)]

	v := pool.Get()
	if v == nil {
		ret := make([]float64, sz, sz)
		return &ret
	}

	return v.(*[]float64)
}

func Put(buf *[]float64) {
	if !isPowerOfTwo(len(*buf)) {
		panic(fmt.Sprintf("can't put buffer with size not a power of 2: %d", len(*buf)))
	}
	pool := pools[bits.TrailingZeros(uint(len(*buf))>>7)]
	pool.Put(buf)
}
