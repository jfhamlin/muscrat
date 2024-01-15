package chart

import (
	"fmt"
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

		xs, ys, smoothYs []float64
		dataMtx          sync.RWMutex

		minSize fyne.Size
	}

	AxisConfig struct {
		Label     string
		Max       float64
		Min       float64
		Clamp     bool
		Log       bool
		Precision int
	}

	// LineChartConfig is a configuration struct for a LineChart.
	LineChartConfig struct {
		X, Y   AxisConfig
		Smooth bool
	}

	lineChartRenderer struct {
		widget *LineChart

		segments []*canvas.Line

		xAxis axisItems
		yAxis axisItems

		objects []fyne.CanvasObject
	}

	axisItems struct {
		text  *canvas.Text
		line  *canvas.Line
		ticks []axisTick
	}

	axisTick struct {
		text *canvas.Text
		line *canvas.Line
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
		if len(lc.xs) == len(ys) {
			xs = lc.xs
		} else {
			xs = make([]float64, len(ys))
		}
		for i := range xs {
			xs[i] = float64(i)
		}
	}
	if len(xs) != len(ys) {
		panic("xs and ys must have the same length")
	}

	var smoothYs []float64
	if lc.config.Smooth {
		smoothYs = make([]float64, len(ys))
		// average of the N points around it
		const N = 2
		for i := range ys {
			var sum float64
			for j := i - N/2; j < i+N/2; j++ {
				if j < 0 || j >= len(ys) {
					continue
				}
				sum += ys[j]
			}
			smoothYs[i] = sum / float64(N)
		}
	}

	lc.dataMtx.Lock()
	lc.xs = xs
	lc.ys = ys
	lc.smoothYs = smoothYs
	lc.dataMtx.Unlock()

	lc.Refresh()
}

func (lc *LineChart) CreateRenderer() fyne.WidgetRenderer {
	lc.ExtendBaseWidget(lc)

	r := &lineChartRenderer{
		widget: lc,
		xAxis: axisItems{
			line: canvas.NewLine(theme.ForegroundColor()),
			text: canvas.NewText(lc.config.X.Label, theme.ForegroundColor()),
		},
		yAxis: axisItems{
			line: canvas.NewLine(theme.ForegroundColor()),
			text: canvas.NewText(lc.config.Y.Label, theme.ForegroundColor()),
		},
	}
	r.Layout(lc.minSize)

	return r
}

////////////////////////////////////////////////////////////////////////////////
// renderer

func (a axisItems) Objects() []fyne.CanvasObject {
	objects := make([]fyne.CanvasObject, 0, len(a.ticks)+2)
	objects = append(objects, a.text, a.line)
	for _, tick := range a.ticks {
		objects = append(objects, tick.text, tick.line)
	}
	return objects
}

func (a axisItems) placeLabels(minVal, maxVal float32, logScale bool, axisSize float32) float32 {
	// evenly distribute labels along the axis
	// find tick text with the largest width
	// move ticks and axis line to the right to make room for the text

	// log scale is log2

	maxWidth := float32(math.Inf(-1))

	numTicks := len(a.ticks)

	viewportStep := axisSize / float32(numTicks)
	offset := viewportStep / 2
	for i, tick := range a.ticks {
		// position the tick
		tick.line.Position1 = fyne.NewPos(0, offset)
		tick.line.Position2 = fyne.NewPos(0, offset)

		// position the text
		tick.text.Move(fyne.NewPos(0, axisSize-offset))

		// set the text
		var val float32
		if logScale {
			val = minVal + (maxVal-minVal)*float32(math.Pow(2, float64(i)))
		} else {
			val = minVal + (maxVal-minVal)*float32(i)/float32(numTicks-1)
		}
		newText := fmt.Sprintf("%.0f", val)
		if newText != tick.text.Text {
			tick.text.Text = newText
			tick.text.Refresh()
		}
		if tick.text.MinSize().Width > maxWidth {
			maxWidth = tick.text.MinSize().Width
		}

		offset += viewportStep
	}
	// now right-align the text
	for _, tick := range a.ticks {
		pos := tick.text.Position()
		tick.text.Move(fyne.NewPos(maxWidth-tick.text.MinSize().Width, pos.Y))
	}

	// position the axis line
	axisX := maxWidth + 2
	a.line.Position1 = fyne.NewPos(axisX, 0)
	a.line.Position2 = fyne.NewPos(axisX, axisSize)

	return axisX
}

func (r *lineChartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 100) // Minimum size of the widget
}

func (r *lineChartRenderer) Layout(size fyne.Size) {
	r.Refresh() // Refresh the drawing when layout changes
}

func (r *lineChartRenderer) Refresh() {
	r.widget.dataMtx.RLock()
	defer r.widget.dataMtx.RUnlock()

	w, h := r.widget.Size().Width, r.widget.Size().Height

	if len(r.widget.xs) <= 1 {
		r.objects = nil
		return
	}

	// updateObjects is set to true when the number of CanvasObjects in
	// the renderer changes.
	var updateObjects bool

	const tickYPadding = 10

	////////////////////////////////////////////////////////////////////////////////
	// Y axis
	{
		// how many ticks? distribute them evenly along widget height with
		// a minimum of 2
		titleWidth := r.yAxis.text.MinSize().Height
		r.yAxis.text.Move(fyne.NewPos(w/2-titleWidth/2, 0))

		textHeight := r.yAxis.text.MinSize().Height
		ticks := int(math.Max(2, float64(h/(textHeight+tickYPadding))))
		if ticks < len(r.yAxis.ticks) {
			updateObjects = true
			r.yAxis.ticks = r.yAxis.ticks[:ticks]
		} else if ticks > len(r.yAxis.ticks) {
			updateObjects = true
			for i := len(r.yAxis.ticks); i < ticks; i++ {
				r.yAxis.ticks = append(r.yAxis.ticks, axisTick{
					text: canvas.NewText("", theme.ForegroundColor()),
					line: canvas.NewLine(theme.ForegroundColor()),
				})
				text := r.yAxis.ticks[i].text
				text.TextSize = 2 * theme.TextSize() / 3
				line := r.yAxis.ticks[i].line
				line.StrokeWidth = 1
			}
		}

		// setting tick positions
		for i := range r.yAxis.ticks {
			// position each tick evenly along the height
			line := r.yAxis.ticks[i].line
			line.Position1 = fyne.NewPos(5, float32(i)*(h/float32(ticks)))
			line.Position2 = fyne.NewPos(10, float32(i)*(h/float32(ticks)))
		}

		r.yAxis.line.Position1 = fyne.NewPos(10, 0)
		r.yAxis.line.Position2 = fyne.NewPos(10, h)
	}
	// Y axis
	////////////////////////////////////////////////////////////////////////////////

	if len(r.segments) != len(r.widget.xs)-1 {
		updateObjects = true
		segments := len(r.widget.xs) - 1
		r.segments = make([]*canvas.Line, segments)
		for i := 0; i < segments; i++ {
			r.segments[i] = canvas.NewLine(theme.ForegroundColor())
			r.segments[i].StrokeWidth = 1
			r.objects = append(r.objects, r.segments[i])
		}
	}

	if updateObjects {
		r.objects = r.objects[:0]
		r.objects = append(r.objects, r.yAxis.Objects()...)
		for _, seg := range r.segments {
			r.objects = append(r.objects, seg)
		}
	}

	xs := r.widget.xs
	ys := r.widget.ys
	if r.widget.config.Smooth {
		ys = r.widget.smoothYs
	}

	maxY := math.Inf(-1)
	minY := math.Inf(1)
	maxX := math.Inf(-1)
	minX := math.Inf(1)
	for i := range ys {
		x := xs[i]
		maxX = math.Max(maxX, x)
		minX = math.Min(minX, x)
		if r.widget.config.X.Clamp {
			maxX = math.Min(maxX, r.widget.config.X.Max)
			minX = math.Max(minX, r.widget.config.X.Min)
		}

		y := ys[i]
		maxY = math.Max(maxY, y)
		minY = math.Min(minY, y)
		if r.widget.config.Y.Clamp {
			maxY = math.Min(maxY, r.widget.config.Y.Max)
			minY = math.Max(minY, r.widget.config.Y.Min)
		}
	}
	// extend bounds to at least configured max/min
	maxX = math.Max(maxX, r.widget.config.X.Max)
	minX = math.Min(minX, r.widget.config.X.Min)
	maxY = math.Max(maxY, r.widget.config.Y.Max)
	minY = math.Min(minY, r.widget.config.Y.Min)

	logMinX := math.Log2(minX)
	logMaxX := math.Log2(maxX)
	logMinY := math.Log2(minY)
	logMaxY := math.Log2(maxY)

	graphX := r.yAxis.placeLabels(float32(minY), float32(maxY), r.widget.config.Y.Log, h)
	graphW := w - graphX

	for i, line := range r.segments {
		var x1, x2, y1, y2 float64

		y1 = ys[i]
		y2 = ys[i+1]
		if r.widget.config.Y.Clamp {
			y1 = math.Max(y1, r.widget.config.Y.Min)
			y1 = math.Min(y1, r.widget.config.Y.Max)

			y2 = math.Max(y2, r.widget.config.Y.Min)
			y2 = math.Min(y2, r.widget.config.Y.Max)
		}
		if r.widget.config.Y.Log {
			y1 = (math.Log2(y1) - logMinY) / (logMaxY - logMinY)
			y2 = (math.Log2(y2) - logMinY) / (logMaxY - logMinY)
		} else {
			y1 = (y1 - minY) / (maxY - minY)
			y2 = (y2 - minY) / (maxY - minY)
		}
		y1 = 1 - y1
		y2 = 1 - y2

		x1 = xs[i]
		x2 = xs[i+1]
		if r.widget.config.X.Clamp {
			x1 = math.Max(x1, r.widget.config.X.Min)
			x1 = math.Min(x1, r.widget.config.X.Max)

			x2 = math.Max(x2, r.widget.config.X.Min)
			x2 = math.Min(x2, r.widget.config.X.Max)
		}
		if r.widget.config.X.Log {
			x1 = (math.Log2(x1) - logMinX) / (logMaxX - logMinX)
			x2 = (math.Log2(x2) - logMinX) / (logMaxX - logMinX)
		} else {
			x1 = (x1 - minX) / (maxX - minX)
			x2 = (x2 - minX) / (maxX - minX)
		}

		line.Position1.X = graphW*float32(x1) + graphX
		line.Position2.X = graphW*float32(x2) + graphX

		line.Position1.Y = h * float32(y1)
		line.Position2.Y = h * float32(y2)
	}
}

func (r *lineChartRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *lineChartRenderer) Destroy() {}
