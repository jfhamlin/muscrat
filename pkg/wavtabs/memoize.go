package wavtabs

import "sync"

type (
	memoKey struct {
		typ string
		res int
	}
)

var (
	memoized = make(map[memoKey]*Table)
	memoLock sync.Mutex
)

func memoize(typ string, res int, fn func() *Table) *Table {
	memoLock.Lock()
	defer memoLock.Unlock()

	key := memoKey{typ, res}
	if tbl, ok := memoized[key]; ok {
		return tbl
	}

	tbl := fn()
	memoized[key] = tbl
	return tbl
}
