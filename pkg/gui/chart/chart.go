package chart

import (
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type (
	// LineChart is a chart that displays a line graph.
	LineChart struct {
		widget.BaseWidget

		xLabel string
		yLabel string
		data   []float64

		minSize fyne.Size
	}

	// LineChartConfig is a configuration struct for a LineChart.
	LineChartConfig struct {
		xLabel string
		yLabel string
		data   []float64
	}

	lineChartRenderer struct {
		lineChart *LineChart

		drawMtx   sync.Mutex
		pltCanvas *vgimg.Canvas
		image     *canvas.Image

		objects []fyne.CanvasObject
	}
)

var (
	_ fyne.WidgetRenderer = (*lineChartRenderer)(nil)
)

func NewLineChart(cfg *LineChartConfig) *LineChart {
	lc := &LineChart{
		data:    cfg.data,
		xLabel:  cfg.xLabel,
		yLabel:  cfg.yLabel,
		minSize: fyne.Size{Width: 300, Height: 264},
	}

	lc.ExtendBaseWidget(lc)
	return lc
}

func (lc *LineChart) SetMinSize(size fyne.Size) {
	lc.minSize = size
}

func (lc *LineChart) Update(data []float64) {
	lc.data = data
	lc.Refresh()
}

func (lc *LineChart) CreateRenderer() fyne.WidgetRenderer {
	lc.ExtendBaseWidget(lc)

	r := &lineChartRenderer{
		lineChart: lc,
	}
	r.Layout(lc.minSize)
	r.objects = []fyne.CanvasObject{r.image}

	return r
}

////////////////////////////////////////////////////////////////////////////////
// Renderer

func (r *lineChartRenderer) draw() {
	p := plot.New()
	p.BackgroundColor = theme.BackgroundColor()

	p.X.Label.Text = r.lineChart.xLabel
	p.X.Label.TextStyle.Color = theme.ForegroundColor()
	p.X.Tick.Label.Color = theme.ForegroundColor()
	p.X.Tick.Color = theme.ForegroundColor()
	p.X.LineStyle.Color = theme.ForegroundColor()

	p.Y.Label.Text = r.lineChart.yLabel
	p.Y.Label.TextStyle.Color = theme.ForegroundColor()
	p.Y.LineStyle.Color = theme.ForegroundColor()
	p.Y.Tick.Label.Color = theme.ForegroundColor()
	p.Y.Tick.Color = theme.ForegroundColor()
	p.Y.Min = -1
	p.Y.Max = 1

	pts := make(plotter.XYs, len(r.lineChart.data))
	for i, v := range r.lineChart.data {
		pts[i].X = float64(i)
		pts[i].Y = v
		p.Y.Min = math.Min(p.Y.Min, v)
		p.Y.Max = math.Max(p.Y.Max, v)
	}

	line, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	line.Color = theme.ForegroundColor()
	p.Add(line)

	p.Draw(draw.New(r.pltCanvas))
}

func (r *lineChartRenderer) Layout(size fyne.Size) {
	r.drawMtx.Lock()
	defer r.drawMtx.Unlock()

	if r.pltCanvas != nil {
		w, h := r.pltCanvas.Size()
		if float32(w) == size.Width && float32(h) == size.Height {
			return
		}
	}

	r.pltCanvas = vgimg.New(font.Length(size.Width), font.Length(size.Height))
	r.image = canvas.NewImageFromImage(r.pltCanvas.Image())
	r.image.FillMode = canvas.ImageFillContain
	r.image.SetMinSize(size)
	r.image.Resize(size)
	r.objects = []fyne.CanvasObject{r.image}
	r.draw()
}

func (r *lineChartRenderer) MinSize() fyne.Size {
	return r.lineChart.minSize
}

func (r *lineChartRenderer) Refresh() {
	r.Layout(r.lineChart.Size())

	r.drawMtx.Lock()
	defer r.drawMtx.Unlock()

	r.draw()

	r.image.Refresh()
}

func (r *lineChartRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *lineChartRenderer) Destroy() {
}
