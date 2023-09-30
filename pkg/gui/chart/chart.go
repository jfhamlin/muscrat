package chart

import (
	"image/png"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type (
	// LineChart is a chart that displays a line graph.
	LineChart struct {
		*canvas.Image

		xLabel    string
		yLabel    string
		data      []float64
		pltCanvas *vgimg.Canvas
	}
)

var (
	_ fyne.CanvasObject = (*LineChart)(nil)
)

func NewLineChart(data []float64) *LineChart {
	lc := &LineChart{
		data:      data,
		pltCanvas: vgimg.New(320, 240),
	}
	lc.draw()
	lc.Image = canvas.NewImageFromImage(lc.pltCanvas.Image())
	lc.FillMode = canvas.ImageFillContain
	return lc
}

func (lc *LineChart) CanvasObject() fyne.CanvasObject {
	return lc.Image
}

func (lc *LineChart) Update(data []float64) {
	lc.data = data
	lc.Refresh()
}

func (lc *LineChart) draw() {
	p := plot.New()
	p.X.Label.Text = lc.xLabel
	p.Y.Label.Text = lc.yLabel

	pts := make(plotter.XYs, len(lc.data))
	for i, v := range lc.data {
		pts[i].X = float64(i)
		pts[i].Y = v
	}

	line, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	p.Add(line)

	p.Draw(draw.New(lc.pltCanvas))

	// save lc.img to file "out.png"
	f, err := os.Create("out.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	png.Encode(f, lc.pltCanvas.Image())
}

func (lc *LineChart) Refresh() {
	lc.draw()
	// set lc.img as the canvas
	lc.Image.Refresh()
}
