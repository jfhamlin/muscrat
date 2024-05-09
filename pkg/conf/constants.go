// Package conf provides static configuration values for the
// application.
package conf

import (
	"math/bits"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jfhamlin/muscrat/internal/pkg/platform"
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

	// SampleFilePaths is the list of paths to directories containing
	// sample files.
	SampleFilePaths = func() []string {
		path := os.Getenv("MUSCRAT_SAMPLE_PATH")
		if path != "" {
			return strings.Split(path, ":")
		}

		resourcesPath := platform.ResourcesPath()
		if resourcesPath == "" {
			return nil
		}

		return []string{filepath.Join(resourcesPath, "samples")}
	}()
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
