// Package conf provides static configuration values for the
// application.
package conf

import (
	"math/bits"
	"os"
	"strconv"

	"github.com/jfhamlin/muscrat/pkg/bufferpool"
)

var (
	// BufferSize is the size of the buffer used for processing one
	// block of samples.
	BufferSize = func() int {
		val := getValueInt("MUSCRAT_BUFFER_SIZE", 128)
		// if not a power of 2, round up to the next power of 2
		if val&(val-1) != 0 {
			leadingZeros := bits.LeadingZeros(uint(val))
			val = 1 << (32 - leadingZeros)
		}
		return clamp(bufferpool.MinSize, bufferpool.MaxSize, val)
	}()

	// SampleRate is the sample rate of the audio system.
	SampleRate = clamp(22050, 192000, getValueInt("MUSCRAT_SAMPLE_RATE", 44100))
)

func clamp(min, max, value int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func getValueInt(envVar string, defaultValue int) int {
	if v := os.Getenv(envVar); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}
