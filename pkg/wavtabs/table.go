//go:generate go run ./internal/gen-tables -o tables.txt
package wavtabs

import (
	"math"
)

const (
	// DefaultResolution is the default resolution for wavetables.
	DefaultResolution = 2048
)

type (
	// Table is a wavetable.
	Table struct {
		// len(tbl) == res + 1, so that tbl[res] == tbl[0]
		tbl []float64
		res int

		bins []complex128
	}
)

// New returns a new wavetable with the given control points.
func New(points []float64) *Table {
	return NewWithWrap(points, points[0])
}

func NewWithWrap(points []float64, wrapVal float64) *Table {
	tbl := make([]float64, len(points)+1)
	copy(tbl, points)
	tbl[len(points)] = wrapVal
	return &Table{
		tbl: tbl,
		res: len(points),
	}
}

func (t *Table) Resolution() int {
	return t.res
}

// Nearest returns the nearest discrete value in a wavetable. The
// function described by the table is assumed to be periodic, with the
// contents describing a single period of the waveform covering [0, 1).
// Values of x outside this range are valid, and will be wrapped to the
// appropriate position in the table.
func (t *Table) Nearest(x float64) float64 {
	x -= math.Floor(x)
	x = x * float64(t.Resolution())
	i := int(x + 0.5)
	return t.tbl[i]
}

// Lerp linearly interpolates between two discrete values in a
// wavetable. The function described by the table is assumed to be
// periodic, with the contents describing a single period of the
// waveform covering [0, 1). Values of x outside this range are valid,
// and will be wrapped to the appropriate position in the table.
func (t *Table) Lerp(x float64) float64 {
	x -= math.Floor(x)
	x = x * float64(t.Resolution())
	i := int(x)
	f := x - float64(i)
	return t.tbl[i] + f*(t.tbl[i+1]-t.tbl[i])
}

func (t *Table) Hermite(x float64) float64 {
	x -= math.Floor(x)
	x = x * float64(t.Resolution())
	i := int(x)
	off := x - float64(i)

	var val0, val1, val2, val3 float64
	if i == 0 {
		val0 = t.tbl[t.Resolution()-1]
	} else {
		val0 = t.tbl[i-1]
	}
	val1 = t.tbl[i]
	val2 = t.tbl[i+1]
	if i == t.Resolution()-1 {
		val3 = t.tbl[0]
	} else {
		val3 = t.tbl[i+2]
	}

	return hermite(off, val0, val1, val2, val3)
}

func hermite(off, val0, val1, val2, val3 float64) float64 {
	slope0 := (val2 - val0) * 0.5
	slope1 := (val3 - val1) * 0.5
	v := val1 - val2
	w := slope0 + v
	a := w + v + slope1
	bNeg := w + a
	stage1 := a*off - bNeg
	stage2 := stage1*off + slope0
	return stage2*off + val1
}
