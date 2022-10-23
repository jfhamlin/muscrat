package plot

import (
	"math"
	"math/bits"
	"math/cmplx"
	"strconv"
	"strings"

	"github.com/jfhamlin/muscrat/internal/pkg/dsp"
)

// DFTHistogramString returns a string with an ASCII spectral
// histogram of the signal with DFT in bins. len(bins) must be a power
// of 2.
func DFTHistogramString(bins []complex128, sampleRate float64, width, height int) string {
	if bits.OnesCount(uint(len(bins))) != 1 {
		panic("len(bins) must be a power of 2")
	}
	// drop bins above the Nyquist frequency
	halfBins := bins[:len(bins)/2+1]

	maxPower := 0.0
	for _, bin := range halfBins {
		maxPower = math.Max(maxPower, cmplx.Abs(bin))
	}

	plotWidth := width - 3 // borders and newline

	const borderHeight = 2
	const labelHeight = 2
	plotHeight := height - borderHeight - labelHeight

	plotBinHeights := make([]int, plotWidth)
	plotBinSourceBins := make([][]int, plotWidth)
	for i := 0; i < plotWidth; i++ {
		// sum the power of all bins that fall into this column
		power := 0.0
		for k := range halfBins {
			if k*plotWidth/len(halfBins) == i {
				power += cmplx.Abs(halfBins[k])
				plotBinSourceBins[i] = append(plotBinSourceBins[i], k)
			}
		}
		// normalize power to [0, 1]
		power /= maxPower
		// map power to [0, plotHeight]
		plotBinHeights[i] = int(power * float64(plotHeight))
	}

	builder := strings.Builder{}
	writeChartHLine(&builder, width)
	builder.WriteByte('\n')

	for i := 0; i < plotHeight; i++ {
		builder.WriteByte('|')
		for j := 0; j < plotWidth; j++ {
			if plotBinHeights[j] >= plotHeight-i {
				builder.WriteByte('#')
			} else {
				builder.WriteByte(' ')
			}
		}
		builder.WriteByte('|')
		builder.WriteByte('\n')
	}

	freqs := dsp.FFTFreqs(sampleRate, len(bins))
	labels, usedValues := xLabelString(freqs, width)

	builder.WriteByte('+')
Outer:
	for i := 0; i < width-3; i++ {
		for _, sourceBin := range plotBinSourceBins[i] {
			if containsFloat64(usedValues, freqs[sourceBin]) {
				builder.WriteByte('.')
				continue Outer
			}
		}
		builder.WriteByte('-')
	}
	builder.WriteByte('+')
	builder.WriteByte('\n')

	builder.WriteString(labels)

	builder.WriteByte('\n')
	builder.WriteString(strings.Repeat(" ", (width-1-len("Hz"))/2))
	builder.WriteString("Hz")

	return builder.String()
}

func xLabelString(labelValues []float64, width int) (string, []float64) {
	labels := make([]string, len(labelValues))
	for i := range labels {
		labels[i] = strconv.FormatFloat(labelValues[i], 'f', -1, 64)
	}

	usedValues := make([]float64, 0, len(labelValues))

	builder := strings.Builder{}
	offset := 0
	for i := 0; i < len(labels); i++ {
		if i == len(labels)-1 {
			builder.WriteString(strings.Repeat(" ", width-1-offset-len(labels[i])))
		} else if i > 0 {
			off := i*(width-1)/(len(labels)-1) - len(labels[i])/2 - offset
			if off <= 0 {
				continue
			}
			offset += off
			builder.WriteString(strings.Repeat(" ", off))
		}
		usedValues = append(usedValues, labelValues[i])
		builder.WriteString(labels[i])
		offset += len(labels[i])
	}
	return builder.String(), usedValues
}

func containsFloat64(slice []float64, value float64) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
