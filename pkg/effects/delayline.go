package effects

import (
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

type (
	// DelayLine is a delay line effect.
	// The delay line is a circular buffer that can be written to and read from.
	DelayLine struct {
		idxMask      int
		buf          []float64
		writePos     int
		readPosInt   int     // Integer part of read position
		readPosFrac  float64 // Fractional part of read position
		sampleRateHz float64
		maxDelay     float64
	}
)

// NewDelayLine creates a new delay line effect.
func NewDelayLine(sampleRateHz int, maxDelay float64) *DelayLine {
	sz := ugen.NextPowerOf2(int(math.Ceil(maxDelay*float64(sampleRateHz) + 1)))

	return &DelayLine{
		buf:          make([]float64, sz),
		idxMask:      sz - 1,
		sampleRateHz: float64(sampleRateHz),
		maxDelay:     maxDelay,
	}
}

func (dl *DelayLine) WriteSample(s float64) {
	dl.buf[dl.writePos] = s
	dl.writePos = (dl.writePos + 1) & dl.idxMask
}

func (dl *DelayLine) ReadSampleN() float64 {
	res := dl.buf[dl.readPosInt&dl.idxMask]
	dl.readPosInt += 1
	return res
}

// ReadSampleL reads a sample from the delay line using linear interpolation.
func (dl *DelayLine) ReadSampleL() float64 {
	idx0 := dl.readPosInt & dl.idxMask
	idx1 := (dl.readPosInt + 1) & dl.idxMask
	frac := dl.readPosFrac
	y0 := dl.buf[idx0]
	y1 := dl.buf[idx1]
	res := y0 + frac*(y1-y0)

	// Increment read position
	dl.readPosInt += 1
	return res
}

// ReadSampleC reads a sample from the delay line using cubic interpolation.
func (dl *DelayLine) ReadSampleC() float64 {
	idx0 := (dl.readPosInt - 1) & dl.idxMask
	idx1 := dl.readPosInt & dl.idxMask
	idx2 := (dl.readPosInt + 1) & dl.idxMask
	idx3 := (dl.readPosInt + 2) & dl.idxMask
	frac := dl.readPosFrac
	x0 := dl.buf[idx0]
	x1 := dl.buf[idx1]
	x2 := dl.buf[idx2]
	x3 := dl.buf[idx3]
	res := ugen.CubInterp(frac, x0, x1, x2, x3)

	// Increment read position
	dl.readPosInt += 1
	return res
}

// SetDelaySeconds sets the delay time in seconds.
func (dl *DelayLine) SetDelaySeconds(delaySec float64) {
	if delaySec > dl.maxDelay {
		delaySec = dl.maxDelay
	}
	if delaySec < 0 {
		delaySec = 0
	}

	delaySamples := delaySec * dl.sampleRateHz
	readPosFloat := float64(dl.writePos) - delaySamples
	bufLen := float64(dl.idxMask + 1)

	// Wrap around the buffer length
	readPosFloat = math.Mod(readPosFloat, bufLen)
	if readPosFloat < 0 {
		readPosFloat += bufLen
	}

	dl.readPosInt = int(math.Floor(readPosFloat)) & dl.idxMask
	dl.readPosFrac = readPosFloat - math.Floor(readPosFloat)
}
