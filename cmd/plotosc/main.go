package main

import (
	"context"
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"sort"
	"strconv"

	"github.com/jfhamlin/muscrat/pkg/osc"
	"github.com/jfhamlin/muscrat/pkg/ugen"

	"gonum.org/v1/gonum/dsp/fourier"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

const (
	minFreq        = 80
	timeVizSamples = 2048
	genSamples     = 44100
)

var (
	oscs = map[string]ugen.UGen{
		"sine": osc.NewSine(),
		"saw":  osc.NewSaw(),
		"sqr":  osc.NewPulse(ugen.WithDefaultDutyCycle(0.5)),
		"tri":  osc.NewTri(),
	}
)

func main() {
	keys := make([]string, 0, len(oscs))
	for k := range oscs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		osc := oscs[k]
		plotOsc(k, osc)
	}
}

func plotOsc(name string, osc ugen.UGen) {
	// plot the output of the oscillator over genSamples samples at 44100 Hz
	// for frequencies from minFreq Hz to minFreq*2^10 Hz, with plot per octave.
	numRows := 5
	plots := make([][]*plot.Plot, 0, numRows)
	for row := 0; row < numRows; row++ {
		freq := minFreq * math.Pow(4, float64(row))
		t, f := plotOscFreq(name, osc, freq)
		plots = append(plots, []*plot.Plot{t, f})
	}
	// combine the plots into a single image
	img := vgimg.New(vg.Points(2000), vg.Points(2000))
	dc := draw.New(img)

	t := draw.Tiles{
		Rows:      5,
		Cols:      2,
		PadX:      vg.Points(5),
		PadY:      vg.Points(5),
		PadTop:    vg.Points(5),
		PadBottom: vg.Points(5),
		PadLeft:   vg.Points(5),
		PadRight:  vg.Points(5),
	}

	// draw the plots onto the image
	canvases := plot.Align(plots, t, dc)
	for row := 0; row < len(plots); row++ {
		for col := 0; col < len(plots[row]); col++ {
			plots[row][col].Draw(canvases[row][col])
		}
	}

	// write the image to a file
	f, err := os.Create(name + ".png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	png := vgimg.PngCanvas{Canvas: img}
	if _, err := png.WriteTo(f); err != nil {
		panic(err)
	}
}

func plotOscFreq(name string, osc ugen.UGen, freq float64) (timePlot *plot.Plot, freqPlot *plot.Plot) {
	// use gonum/plot to plot the output of the oscillator over genSamples
	// samples at 44100 Hz for the given frequency.
	//
	// plot both the time domain and frequency domain representations.

	const sampleRate = 44100

	freqSlice := make([]float64, genSamples)
	for i := 0; i < genSamples; i++ {
		freqSlice[i] = freq
	}

	samples := make([]float64, genSamples)
	cfg := ugen.SampleConfig{
		SampleRateHz: sampleRate,
		InputSamples: map[string][]float64{
			"w": freqSlice,
		},
	}
	osc.Gen(context.Background(), cfg, samples)

	return plotTimeDomain(name, freq, sampleRate, samples), plotFreqDomain(name, freq, sampleRate, samples)
}

func plotTimeDomain(name string, freq, sampleRate float64, samples []float64) *plot.Plot {
	p := plot.New()

	p.Title.Text = fmt.Sprintf("%s: %.1f Hz", name, freq)
	p.X.Label.Text = "time"
	p.Y.Label.Text = "amplitude"

	pts := make(plotter.XYs, timeVizSamples)
	for i := range pts {
		// X is time in seconds
		pts[i].X = float64(i) / sampleRate
		pts[i].Y = samples[i]
	}

	l, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	p.Add(l)

	return p
}

func plotFreqDomain(name string, freq, sampleRate float64, samples []float64) *plot.Plot {
	// first, calculate the FFT of the samples after applying a window
	// function to reduce spectral leakage.
	//
	// we'll use a Hann window, which is a cosine function that goes from
	// 0 to 1 over the length of the sample.
	//
	// the window function is applied to the samples before calculating
	// the FFT.
	//
	// the FFT is calculated using the GoNum package.
	//
	// the FFT is then converted to a power spectrum, which is the
	// magnitude of the FFT squared.
	//
	// the power spectrum is then converted to decibels, which is 10 *
	// log10(power).
	//
	// the power spectrum is then plotted against the frequency bins
	// returned by the FFT.

	// apply the Hann window to the samples
	window := make([]float64, len(samples))
	for i := range window {
		window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(len(window)-1)))
		samples[i] *= window[i]
	}

	// calculate the FFT of the samples. note that only the first half
	// of the FFT is returned (genSamples/2 + 1).
	fft := fourier.NewFFT(len(samples))
	bins := fft.Coefficients(nil, samples)

	// convert the FFT to a power spectrum
	power := make([]float64, len(bins))
	for i := range power {
		power[i] = cmplx.Abs(bins[i])
	}

	// convert the power spectrum to decibels
	db := make([]float64, len(power))
	for i := range db {
		db[i] = 10 * math.Log10(power[i])
	}

	// create the plot
	p := plot.New()

	p.Title.Text = fmt.Sprintf("%s: %.1f Hz", name, freq)
	p.X.Label.Text = "frequency"
	p.X.Scale = plot.LogScale{}
	p.X.Tick.Marker = LogTicks{Prec: 1}
	p.Y.Label.Text = "power (dB)"

	pts := make(plotter.XYs, len(db))
	for i := range pts {
		// X is frequency in Hz
		pts[i].X = float64(i) * sampleRate / genSamples
		pts[i].Y = power[i]
		//pts[i].Y = db[i]
	}
	const minPlotFreq = minFreq / 2
	pts = pts[int(minPlotFreq*genSamples/sampleRate):]

	l, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	p.Add(l)

	return p
}

// LogTicks is suitable for the Tick.Marker field of an Axis,
// it returns tick marks suitable for a log-scale axis.
type LogTicks struct {
	// Prec specifies the precision of tick rendering
	// according to the documentation for strconv.FormatFloat.
	Prec int
}

// Ticks returns Ticks in a specified range
func (t LogTicks) Ticks(min, max float64) []plot.Tick {
	if min <= 0 || max <= 0 {
		panic("Values must be greater than 0 for a log scale.")
	}

	val := math.Pow10(int(math.Log10(min)))
	max = math.Pow10(int(math.Ceil(math.Log10(max))))
	var ticks []plot.Tick
	for val < max {
		for i := 1; i < 10; i++ {
			if i == 1 {
				ticks = append(ticks, plot.Tick{Value: val, Label: formatFloatTick(val, t.Prec)})
			}
			tickVal := val * float64(i)
			if i == 2 || i == 5 {
				ticks = append(ticks, plot.Tick{Value: tickVal, Label: formatFloatTick(tickVal, t.Prec)})
			} else {
				ticks = append(ticks, plot.Tick{Value: val * float64(i)})
			}
		}
		val *= 10
	}
	ticks = append(ticks, plot.Tick{Value: val, Label: formatFloatTick(val, t.Prec)})

	return ticks
}

func formatFloatTick(v float64, prec int) string {
	return strconv.FormatFloat(v, 'f', prec, 64)
}
