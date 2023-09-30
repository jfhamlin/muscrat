package gui

import (
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"

	"github.com/jfhamlin/muscrat/pkg/gui/chart"
)

func NewMainWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("Muscrat")

	logo := LogoImage()
	logo.SetMinSize(fyne.NewSize(100, 100))

	// Generate some sample data
	var data []float64
	for i := 0; i < 360; i += 10 {
		data = append(data, math.Sin(float64(i)*math.Pi/180))
	}
	wave := chart.NewLineChart(data)
	wave.SetMinSize(fyne.NewSize(320, 240))
	go func() {
		ticker := time.NewTicker(time.Second / 30)
		angle := 0.0
		last := time.Now()
		for range ticker.C {
			now := time.Now()
			diff := now.Sub(last)
			angle += diff.Seconds() * 180
			for i := range data {
				angleOffset := float64(i) * 10
				data[i] = math.Sin((angle + angleOffset) * math.Pi / 180)
			}
			last = now
			wave.Update(data)
		}
	}()

	w.SetContent(container.NewVBox(
		logo,
		wave.CanvasObject(),
	))
	return w
}
