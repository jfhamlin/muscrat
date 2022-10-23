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
	logRange  bool
	minFreq   float64
	maxFreq   float64
}

// WithLogDomain returns a PlotOption that causes the plot to be
// rendered with a logarithmic x axis.
func WithLogDomain() PlotOption {
	return func(o *plotOptions) {
		o.logDomain = true
	}
}

// WithLogRange returns a PlotOption that causes the plot to be
// rendered with a logarithmic y axis.
func WithLogRange() PlotOption {
	return func(o *plotOptions) {
		o.logRange = true
	}
}

// WithMinFreq returns a PlotOption that causes the plot to be
// rendered omitting frequencies below minFreq.
func WithMinFreq(minFreq float64) PlotOption {
	return func(o *plotOptions) {
		o.minFreq = minFreq
	}
}

// WithMaxFreq returns a PlotOption that causes the plot to be
// rendered omitting frequencies above maxFreq.
func WithMaxFreq(maxFreq float64) PlotOption {
	return func(o *plotOptions) {
		o.maxFreq = maxFreq
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

	plotWidth := width - 3 // borders and newline

	const borderHeight = 2
	const labelHeight = 2
	plotHeight := height - borderHeight - labelHeight

	plotBinHeights := make([]int, plotWidth)
	plotBinLabelValues := make([][]float64, plotWidth)
	freqs := dsp.FFTFreqs(sampleRate, len(bins))
	if len(freqs) != len(halfBins) {
		panic("len(freqs) != len(halfBins)")
	}

	if o.maxFreq > 0 {
		maxBin := int(o.maxFreq / freqs[1])
		if maxBin < len(halfBins)-1 {
			halfBins = halfBins[:maxBin+1]
			freqs = freqs[:maxBin+1]
		}
	}
	if o.minFreq > 0 {
		minBin := int(o.minFreq / freqs[1])
		halfBins = halfBins[minBin:]
		freqs = freqs[minBin:]
	}

	plotPower := make([]float64, plotWidth)
	if !o.logDomain {
		freqStep := freqs[1] - freqs[0]
		plotFreqRange := freqs[len(freqs)-1] - freqs[0] + freqStep
		plotFreqStep := plotFreqRange / float64(plotWidth)
		halfBinCursor := 0
		for i := 0; i < plotWidth; i++ {
			colMaxFreq := freqs[0] + float64(i+1)*plotFreqStep
			// sum the power of all bins that fall into this column
			power := 0.0
			for halfBinCursor < len(freqs) && freqs[halfBinCursor] < colMaxFreq {
				plotBinLabelValues[i] = append(plotBinLabelValues[i], freqs[halfBinCursor])
				power += cmplx.Abs(halfBins[halfBinCursor])
				halfBinCursor++
			}
			plotPower[i] = power
		}
	} else {
		logFreqs := make([]float64, len(freqs))
		minLogFreq := math.Log(freqs[1] / 2)
		for i := range logFreqs {
			if freqs[i] == 0 {
				logFreqs[i] = minLogFreq
			} else {
				logFreqs[i] = math.Log2(freqs[i])
			}
		}

		maxFreq := sampleRate/2 + freqs[1] - freqs[0]
		if o.maxFreq > 0 {
			maxFreq = o.maxFreq + freqs[1] - freqs[0]
		}
		if o.minFreq > 0 {
			minLogFreq = math.Log2(o.minFreq)
		}
		plotLogFreqRange := math.Log2(maxFreq) - minLogFreq
		plotLogFreqStep := plotLogFreqRange / float64(plotWidth)
		halfBinCursor := 0
		for i := 0; i < plotWidth; i++ {
			colMaxLogFreq := minLogFreq + float64(i+1)*plotLogFreqStep
			// sum the power of all bins that fall into this column
			power := 0.0
			for halfBinCursor < len(logFreqs) && logFreqs[halfBinCursor] < colMaxLogFreq {
				plotBinLabelValues[i] = append(plotBinLabelValues[i], freqs[halfBinCursor])
				power += cmplx.Abs(halfBins[halfBinCursor])
				halfBinCursor++
			}
			plotPower[i] = power

			if len(plotBinLabelValues[i]) == 0 {
				plotBinLabelValues[i] = append(plotBinLabelValues[i], math.Pow(2, colMaxLogFreq))
			}
		}
	}

	maxPower := 0.0
	for i := range plotPower {
		if o.logRange {
			plotPower[i] = math.Log2(plotPower[i] + 1)
		}
		power := plotPower[i]
		if power > maxPower {
			maxPower = power
		}
	}

	for i, power := range plotPower {
		// normalize plotPower to [0, 1]
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

// tryFitLabels tries to fit the labels into a width len(labels)+2
// line such that each label is centered on its column. If the labels
// are too long to fit, it returns false.
func tryFitLabels(labels []string) (string, bool) {
	var builder strings.Builder
	for i, label := range labels {
		if len(label) == 0 {
			continue
		}

		col := i + 1
		targetOffset := col - len(label)/2
		if targetOffset < 0 {
			targetOffset = 0
		}
		if builder.Len() < targetOffset {
			builder.WriteString(strings.Repeat(" ", targetOffset-builder.Len()))
		} else if i > 0 && builder.Len() >= targetOffset {
			return "", false
		}
		builder.WriteString(label)
	}
	if builder.Len() > len(labels)+2 {
		return "", false
	}
	var tickBuilder strings.Builder
	tickBuilder.WriteByte('+')
	for i := 0; i < len(labels); i++ {
		if len(labels[i]) == 0 {
			tickBuilder.WriteByte('-')
		} else {
			tickBuilder.WriteByte('.')
		}
	}
	tickBuilder.WriteByte('+')
	tickBuilder.WriteByte('\n')
	tickBuilder.WriteString(builder.String())

	return tickBuilder.String(), true
}

func xLabelString(labelValues [][]float64, width int) string {
	if len(labelValues) != width-3 {
		panic("len(labelValues) != width-3")
	}

	labels := make([]string, len(labelValues))
	for i := range labels {
		if len(labelValues[i]) > 0 {
			labels[i] = formatFloat64(labelValues[i][0])
		}
	}

	for step := 1; step < len(labels); step++ {
		tryLabels := make([]string, len(labels))
		for i := range tryLabels {
			if i%step == 0 {
				tryLabels[i] = labels[i]
			}
		}
		if tryString, ok := tryFitLabels(tryLabels); ok {
			return tryString
		}
	}
	return strings.Repeat("-", len(labels)+2)
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
