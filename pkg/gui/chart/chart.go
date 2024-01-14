package chart

import (
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	minPixelsPerLabel = 50
)

type (
	// LineChart is a chart that displays a line graph.
	LineChart struct {
		widget.BaseWidget

		config LineChartConfig

		xs, ys  []float64
		dataMtx sync.RWMutex

		minSize fyne.Size
	}

	AxisConfig struct {
		Label string
		Max   float64
		Min   float64
		Clamp bool
		Log   bool
	}

	// LineChartConfig is a configuration struct for a LineChart.
	LineChartConfig struct {
		X, Y AxisConfig
	}

	lineChartRenderer struct {
		widget *LineChart

		segments []*canvas.Line

		objects []fyne.CanvasObject
	}

	log2Scale struct{}

	log2Ticks struct {
		Prec  int
		Width float64
	}
)

func NewLineChart(cfg LineChartConfig) *LineChart {
	lc := &LineChart{
		config:  cfg,
		minSize: fyne.Size{Width: 300, Height: 264},
	}

	lc.ExtendBaseWidget(lc)
	return lc
}

func (lc *LineChart) SetMinSize(size fyne.Size) {
	lc.minSize = size
}

func (lc *LineChart) SetData(xs, ys []float64) {
	if xs == nil {
		xs = make([]float64, len(ys))
	}
	if len(xs) != len(ys) {
		panic("xs and ys must have the same length")
	}

	lc.dataMtx.Lock()
	lc.xs = xs
	lc.ys = ys
	lc.dataMtx.Unlock()

	lc.Refresh()
}

func (lc *LineChart) CreateRenderer() fyne.WidgetRenderer {
	lc.ExtendBaseWidget(lc)

	r := &lineChartRenderer{
		widget: lc,
	}
	r.Layout(lc.minSize)

	return r
}

func (r *lineChartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 100) // Minimum size of the widget
}

func (r *lineChartRenderer) Layout(size fyne.Size) {
	r.Refresh() // Refresh the drawing when layout changes
}

func (r *lineChartRenderer) Refresh() {
	if len(r.widget.xs) <= 1 {
		r.objects = nil
		return
	}
	if len(r.segments) != len(r.widget.xs)-1 {
		segments := len(r.widget.xs) - 1
		r.segments = make([]*canvas.Line, segments)
		r.objects = r.objects[:0]
		for i := 0; i < segments; i++ {
			r.segments[i] = canvas.NewLine(theme.ForegroundColor())
			r.segments[i].StrokeWidth = 1
			r.objects = append(r.objects, r.segments[i])
		}
	}
	max := math.Inf(-1)
	min := math.Inf(1)
	for _, y := range r.widget.ys {
		if y > max {
			max = y
		}
		if y < min {
			min = y
		}
	}

	w, h := r.widget.Size().Width, r.widget.Size().Height
	for i, line := range r.segments {
		y1 := float64(h) - (r.widget.ys[i]-min)/(max-min)*float64(h)
		y2 := float64(h) - (r.widget.ys[i+1]-min)/(max-min)*float64(h)

		// ignore widget xs; position equally spaced
		x1 := float64(i) / float64(len(r.widget.xs)-1) * float64(w)
		x2 := float64(i+1) / float64(len(r.widget.xs)-1) * float64(w)

		line.Position1.X = float32(x1)
		line.Position1.Y = float32(y1)
		line.Position2.X = float32(x2)
		line.Position2.Y = float32(y2)
	}
}

func (r *lineChartRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *lineChartRenderer) Destroy() {}

// Renderer

// func (r *lineChartRenderer) drawAxis(xConfig, yConfig AxisConfig) []fyne.CanvasObject {
// 	objs := []fyne.CanvasObject{}

// 	width := r.lineChart.Size().Width
// 	height := r.lineChart.Size().Height

// 	// Draw X-axis
// 	xAxis := canvas.NewLine(theme.ForegroundColor())
// 	xAxis.StrokeWidth = 2
// 	xAxis.Position1 = fyne.NewPos(0, height-20)
// 	xAxis.Position2 = fyne.NewPos(width, height-20)
// 	objs = append(objs, xAxis)

// 	// Draw Y-axis
// 	yAxis := canvas.NewLine(theme.ForegroundColor())
// 	yAxis.StrokeWidth = 2
// 	yAxis.Position1 = fyne.NewPos(20, 0)
// 	yAxis.Position2 = fyne.NewPos(20, height)
// 	objs = append(objs, yAxis)

// 	// Draw labels (basic example, needs more work for proper positioning and formatting)
// 	xLabel := canvas.NewText(xConfig.Label, theme.ForegroundColor())
// 	xLabel.TextSize = 12
// 	xLabel.Move(fyne.NewPos(width/2, height-10))
// 	objs = append(objs, xLabel)

// 	yLabel := canvas.NewText(yConfig.Label, theme.ForegroundColor())
// 	yLabel.TextSize = 12
// 	yLabel.Move(fyne.NewPos(10, height/2))
// 	objs = append(objs, yLabel)

// 	return objs
// }

// func (r *lineChartRenderer) drawLineChart(xConfig, yConfig AxisConfig) []fyne.CanvasObject {
// 	r.lineChart.dataMtx.RLock()
// 	defer r.lineChart.dataMtx.RUnlock()

// 	objs := []fyne.CanvasObject{}

// 	width := r.lineChart.Size().Width - 40   // 20px padding on each side
// 	height := r.lineChart.Size().Height - 40 // 20px padding on each side
// 	xMin, xMax := xConfig.Min, xConfig.Max
// 	yMin, yMax := yConfig.Min, yConfig.Max

// 	for i := 0; i < len(r.lineChart.ys)-1; i++ {
// 		x1 := normalize(r.lineChart.xs[i], xMin, xMax) * width
// 		y1 := (1 - normalize(r.lineChart.ys[i], yMin, yMax)) * height
// 		x2 := normalize(r.lineChart.xs[i+1], xMin, xMax) * width
// 		y2 := (1 - normalize(r.lineChart.ys[i+1], yMin, yMax)) * height

// 		line := canvas.NewLine(theme.ForegroundColor())
// 		line.Position1 = fyne.NewPos(x1+20, y1+20)
// 		line.Position2 = fyne.NewPos(x2+20, y2+20)
// 		line.StrokeWidth = 2
// 		objs = append(objs, line)
// 	}

// 	return objs
// }

// func normalize(value, min, max float64) float32 {
// 	return float32((value - min) / (max - min))
// }

// func (r *lineChartRenderer) Layout(size fyne.Size) {
// 	// Redraw the chart with the new size
// 	r.Refresh()
// }

// func (r *lineChartRenderer) Refresh() {
// 	// Clear existing objects
// 	objects := r.objects[:0]

// 	// Redraw the axes and the line chart
// 	objects = append(objects, r.drawAxis(r.lineChart.config.X, r.lineChart.config.Y)...)
// 	objects = append(objects, r.drawLineChart(r.lineChart.config.X, r.lineChart.config.Y)...)
// 	r.objects = objects

// 	// Update the canvas objects
// 	canvas.Refresh(r.lineChart)
// }

// func (r *lineChartRenderer) MinSize() fyne.Size {
// 	return r.lineChart.minSize
// }

// func (r *lineChartRenderer) Objects() []fyne.CanvasObject {
// 	return r.objects
// }

// func (r *lineChartRenderer) Destroy() {
// 	// Cleanup resources if any
// }
