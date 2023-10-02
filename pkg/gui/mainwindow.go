package gui

import (
	"math"
	"math/cmplx"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gonum.org/v1/gonum/dsp/fourier"

	"github.com/jfhamlin/muscrat/pkg/gui/chart"
	"github.com/jfhamlin/muscrat/pkg/gui/meter"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

const (
	sampleBufferSize = 4096 * 2
	sampleRate       = 44100
)

type (
	// MainWindow is the main window of the application.
	MainWindow struct {
		fyne.Window

		oscilloscope *chart.LineChart
		spectrogram  *chart.LineChart

		unsub func()
	}

	circularBuffer struct {
		buffer []float64
		index  int
	}

	slider struct {
		slider *widget.Slider
		value  *atomic.Value
	}
)

// NewMainWindow creates a new main window.
func NewMainWindow(a fyne.App) *MainWindow {
	w := a.NewWindow("muscrat")

	logo := LogoImage()
	logo.SetMinSize(fyne.NewSize(100, 100))

	osc := chart.NewLineChart(chart.LineChartConfig{
		Y: chart.AxisConfig{
			Label: "Amplitude",
			Min:   -1,
			Max:   1,
		},
	})
	spect := chart.NewLineChart(chart.LineChartConfig{
		X: chart.AxisConfig{
			Label: "Frequency (Hz)",
			Log:   true,
			Min:   20,
			Max:   sampleRate / 2,
			Clamp: true,
		},
		Y: chart.AxisConfig{
			Label: "Power (dB)",
			Max:   0,
			Min:   -100,
			Clamp: true,
		},
	})

	rateSlider := newSlider(1, 30, 15)
	rateSlider.slider.Step = 1

	volumeMeter := meter.NewVolume(-60, 0)
	volumeMeter.SetMinSize(fyne.NewSize(10, 400))

	scopes := container.NewVBox(
		osc,
		spect,
		rateSlider.slider,
	)
	logoMeter := container.NewVBox(
		logo,
		volumeMeter,
	)
	contents := container.NewHBox(
		logoMeter,
		scopes,
	)
	w.SetContent(contents)

	buffer := &circularBuffer{buffer: make([]float64, sampleBufferSize)}

	// buffer used to keep a linear view of the circular buffer
	readBuffer := make([]float64, len(buffer.buffer))

	lastUpdateTime := time.Now()
	unsub := pubsub.Subscribe("samples", func(evt string, data any) {
		samples := data.([]float64)
		buffer.Append(samples)

		buffer.Get(readBuffer)
		volumeMeter.SetValues(rmsDBPeak(readBuffer))

		if time.Since(lastUpdateTime) < time.Second/time.Duration(rateSlider.Value()) {
			return
		}
		lastUpdateTime = time.Now()
		osc.SetData(nil, readBuffer[:len(readBuffer)/4])
		spect.SetData(fft(readBuffer))
	})

	return &MainWindow{
		Window:       w,
		oscilloscope: osc,
		spectrogram:  spect,
		unsub:        unsub,
	}
}

func (b *circularBuffer) Append(v []float64) {
	for _, s := range v {
		b.buffer[b.index] = s
		b.index = (b.index + 1) % len(b.buffer)
	}
}

func (b *circularBuffer) Get(out []float64) {
	copy(out, b.buffer[b.index:])
	copy(out[len(b.buffer)-b.index:], b.buffer[:b.index])
}

func newSlider(min, max, def float64) *slider {
	s := &slider{
		slider: widget.NewSlider(min, max),
		value:  &atomic.Value{},
	}
	s.value.Store(def)
	s.slider.SetValue(def)
	s.slider.OnChanged = func(v float64) {
		s.value.Store(v)
	}
	return s
}

func (s *slider) Value() float64 {
	return s.value.Load().(float64)
}

func fft(samples []float64) (freqs, powerDB []float64) {
	// apply the Hann window to the samples
	window := make([]float64, len(samples))
	windowSum := 0.0
	for i := range window {
		window[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(len(window)-1)))
		windowSum += window[i]
		samples[i] *= window[i]
	}

	// calculate the FFT of the samples. note that only the first half
	// of the FFT is returned (len(samples)/2 + 1).
	fft := fourier.NewFFT(len(samples))
	bins := fft.Coefficients(nil, samples)

	// https://dsp.stackexchange.com/questions/32076/fft-to-spectrum-in-decibel

	// convert the FFT to a power spectrum
	power := make([]float64, len(bins))
	for i := range power {
		power[i] = cmplx.Abs(bins[i]) * 2 / windowSum
	}

	// convert the power spectrum to decibels
	db := make([]float64, len(power))
	for i := range db {
		db[i] = 20 * math.Log10(power[i])
	}

	freqs = make([]float64, len(db))
	for i := range freqs {
		freqs[i] = float64(i) * sampleRate / float64(len(samples))
	}
	freqs[0] += 0.0001 // avoid log(0)

	return freqs, db
}

func rmsPeak(samples []float64) (rms, peak float64) {
	sum := 0.0
	for _, s := range samples {
		sum += s * s
		if math.Abs(s) > peak {
			peak = math.Abs(s)
		}
	}
	return math.Sqrt(sum / float64(len(samples))), peak
}

func rmsDBPeak(samples []float64) (db, peak float64) {
	rms, peak := rmsPeak(samples)
	return 20 * math.Log10(rms), peak
}
