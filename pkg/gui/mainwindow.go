package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

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
)

func NewMainWindow(a fyne.App) *MainWindow {
	w := a.NewWindow("Muscrat")

	logo := LogoImage()
	logo.SetMinSize(fyne.NewSize(100, 100))

	osc := chart.NewLineChart(&chart.LineChartConfig{})
	spect := chart.NewLineChart(&chart.LineChartConfig{})

	w.SetContent(container.NewVBox(
		logo,
		osc,
		spect,
	))

	unsub := pubsub.Subscribe("samples", func(evt string, data any) {
		samples := data.([]float64)
		cpy := make([]float64, len(samples))
		copy(cpy, samples)
		osc.Update(cpy)
		spect.Update(cpy)
	})

	return &MainWindow{
		Window:       w,
		oscilloscope: osc,
		spectrogram:  spect,
		unsub:        unsub,
	}
}
