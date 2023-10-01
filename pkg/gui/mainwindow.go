package gui

import (
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/jfhamlin/muscrat/pkg/gui/chart"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
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
)

// NewMainWindow creates a new main window.
func NewMainWindow(a fyne.App) *MainWindow {
	w := a.NewWindow("Muscrat")

	logo := LogoImage()
	logo.SetMinSize(fyne.NewSize(100, 100))

	osc := chart.NewLineChart(&chart.LineChartConfig{})
	spect := chart.NewLineChart(&chart.LineChartConfig{})

	updateRate := &atomic.Int32{}
	updateRate.Store(15)
	rateSlider := widget.NewSlider(1, 30)
	rateSlider.Step = 1
	rateSlider.SetValue(float64(updateRate.Load()))
	rateSlider.OnChanged = func(v float64) {
		updateRate.Store(int32(v))
	}

	contents := container.NewVBox(
		logo,
		osc,
		spect,
		rateSlider,
	)
	w.SetContent(contents)

	buffer := &circularBuffer{buffer: make([]float64, 1024)}

	lastUpdateTime := time.Now()
	unsub := pubsub.Subscribe("samples", func(evt string, data any) {
		buffer.Append(data.([]float64))
		if time.Since(lastUpdateTime) < time.Second/time.Duration(updateRate.Load()) {
			return
		}
		lastUpdateTime = time.Now()

		samples := buffer.Get()
		osc.Update(samples)
		spect.Update(samples)
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

func (b *circularBuffer) Get() []float64 {
	out := make([]float64, len(b.buffer))
	copy(out, b.buffer[b.index:])
	copy(out[len(b.buffer)-b.index:], b.buffer[:b.index])
	return out
}
