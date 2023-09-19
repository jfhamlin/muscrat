//go:generate go run ./internal/gen-tables -o tables.txt
package wavtabs

import (
	"math"

	"github.com/jfhamlin/muscrat/pkg/interp"
)

const (
	// DefaultResolution is the default resolution for wavetables.
	DefaultResolution = 2048
)

type (
	// Table is a wavetable.
	Table struct {
		// len(tbl) == res + 1, so that tbl[res] == tbl[0]
		tbl atomicTable

		bins []complex128

		bandLimitedTbls  []atomicTable
		bandLimitedFreqs []float64
	}

	// atomicTable is a wavetable without band-limited interpolation.
	atomicTable []float64
)

// New returns a new wavetable with the given control points.
func New(points []float64) *Table {
	return NewWithWrap(points, points[0])
}

func NewWithWrap(points []float64, wrapVal float64) *Table {
	tbl := make([]float64, len(points)+1)
	copy(tbl, points)
	tbl[len(points)] = wrapVal
	t := &Table{
		tbl: tbl,
	}
	t.genBandLimited()

	return t
}

func (t *Table) Resolution() int {
	return len(t.tbl) - 1
}

// Nearest returns the nearest discrete value in a wavetable. The
// function described by the table is assumed to be periodic, with the
// contents describing a single period of the waveform covering [0, 1).
// Values of x outside this range are valid, and will be wrapped to the
// appropriate position in the table.
func (t *Table) Nearest(x float64) float64 {
	return t.tbl.Nearest(x)
}

// Lerp linearly interpolates between two discrete values in a
// wavetable. The function described by the table is assumed to be
// periodic, with the contents describing a single period of the
// waveform covering [0, 1). Values of x outside this range are valid,
// and will be wrapped to the appropriate position in the table.
func (t *Table) Lerp(x float64) float64 {
	return t.tbl.Lerp(x)
}

func (t *Table) Hermite(x float64) float64 {
	return t.tbl.Hermite(x)
}

////////////////////////////////////////////////////////////////////////////////
// atomicTable

func (at atomicTable) Resolution() int {
	return len(at) - 1
}

func (at atomicTable) Nearest(x float64) float64 {
	x -= math.Floor(x)
	x = x * float64(at.Resolution())
	i := int(x + 0.5)
	return at[i]
}

func (at atomicTable) Lerp(x float64) float64 {
	x -= math.Floor(x)
	x = x * float64(at.Resolution())
	i := int(x)
	f := x - float64(i)
	return at[i] + f*(at[i+1]-at[i])
}

func (at atomicTable) Hermite(x float64) float64 {
	res := at.Resolution()

	x -= math.Floor(x)
	x = x * float64(res)
	i := int(x)
	off := x - float64(i)

	var val0, val1, val2, val3 float64
	if i == 0 {
		val0 = at[res-1]
	} else {
		val0 = at[i-1]
	}
	val1 = at[i]
	val2 = at[i+1]
	if i == res-1 {
		val3 = at[0]
	} else {
		val3 = at[i+2]
	}

	return interp.Hermite(off, val0, val1, val2, val3)
}
