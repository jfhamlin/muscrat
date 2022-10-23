package plot

import (
	"fmt"
	"math"
	"math/bits"
	"math/cmplx"
	"strings"

	"github.com/jfhamlin/muscrat/internal/pkg/dsp"
)

// PlotOption is an option for a plot.
type PlotOption func(*plotOptions)

type plotOptions struct {
	logDomain bool
}

// WithLogDomain returns a PlotOption that causes the plot to be
// rendered with a logarithmic x axis.
func WithLogDomain() PlotOption {
	return func(o *plotOptions) {
		o.logDomain = true
	}
}

// DFTHistogramString returns a string with an ASCII spectral
// histogram of the signal with DFT in bins. len(bins) must be a power
// of 2.
func DFTHistogramString(bins []complex128, sampleRate float64, width, height int, opts ...PlotOption) string {
	if bits.OnesCount(uint(len(bins))) != 1 {
		panic("len(bins) must be a power of 2")
	}

	o := plotOptions{}
	for _, opt := range opts {
		opt(&o)
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
	plotBinLabelValues := make([]float64, plotWidth)
	freqs := dsp.FFTFreqs(sampleRate, len(bins))
	for i := 0; i < plotWidth; i++ {
		// sum the power of all bins that fall into this column
		power := 0.0
		for k := range halfBins {
			if k*plotWidth/len(halfBins) == i {
				power += cmplx.Abs(halfBins[k])
				plotBinLabelValues[i] = freqs[k]
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

	builder.WriteString(xLabelString(plotBinLabelValues, width))

	builder.WriteByte('\n')
	builder.WriteString(strings.Repeat(" ", (width-1-len("Hz"))/2))
	builder.WriteString("Hz")

	return builder.String()
}

func xLabelString(labelValues []float64, width int) string {
	if len(labelValues) != width-3 {
		panic("len(labelValues) != width-3")
	}

	labels := make([]string, len(labelValues))
	for i := range labels {
		labels[i] = formatFloat64(labelValues[i])
	}

	usedValues := make([]float64, 0, len(labelValues))

	labelBuilder := strings.Builder{}
	offset := 0
	for i := 0; i < len(labels); i++ {
		if i == len(labels)-1 {
			padding := width - 1 - offset - len(labels[i])
			if padding < 0 {
				padding = 0
			}
			labelBuilder.WriteString(strings.Repeat(" ", padding))
		} else if i > 0 {
			off := i*(width-1)/(len(labels)-1) - len(labels[i])/2 - offset
			if off <= 0 {
				continue
			}
			offset += off
			labelBuilder.WriteString(strings.Repeat(" ", off))
		}
		usedValues = append(usedValues, labelValues[i])
		labelBuilder.WriteString(labels[i])
		offset += len(labels[i])
	}

	builder := strings.Builder{}
	builder.WriteByte('+')
Outer:
	for i := 0; i < width-3; i++ {
		if containsFloat64(usedValues, labelValues[i]) {
			builder.WriteByte('.')
			continue Outer
		}
		builder.WriteByte('-')
	}
	builder.WriteByte('+')
	builder.WriteByte('\n')
	builder.WriteString(labelBuilder.String())

	return builder.String()
}

func containsFloat64(slice []float64, value float64) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func formatFloat64(x float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", x), "0"), ".")
}
