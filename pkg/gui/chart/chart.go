package chart

import (
	"math"
	"strconv"
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
		lineChart *LineChart

		drawMtx   sync.Mutex
		pltCanvas *vgimg.Canvas
		image     *canvas.Image

		objects []fyne.CanvasObject
	}

	log2Scale struct{}

	log2Ticks struct {
		Prec int
	}
)

var (
	_ fyne.WidgetRenderer = (*lineChartRenderer)(nil)
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
	if xs != nil && len(xs) != len(ys) {
		panic("xs and ys must have the same length")
	}

	newXs := lc.xs
	newYs := lc.ys
	if len(xs) != len(lc.xs) {
		if xs == nil {
			newXs = nil
		} else {
			newXs = make([]float64, len(xs))
		}
	}
	if len(ys) != len(lc.ys) {
		newYs = make([]float64, len(ys))
	}
	copy(newXs, xs)
	copy(newYs, ys)

	lc.dataMtx.Lock()
	lc.xs = newXs
	lc.ys = newYs
	lc.dataMtx.Unlock()

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

	xConfig, yConfig := r.lineChart.config.X, r.lineChart.config.Y

	p.X.Label.Text = xConfig.Label
	p.X.Label.TextStyle.Color = theme.ForegroundColor()
	p.X.Tick.Label.Color = theme.ForegroundColor()
	p.X.Tick.Color = theme.ForegroundColor()
	p.X.LineStyle.Color = theme.ForegroundColor()
	p.X.Min = xConfig.Min
	p.X.Max = xConfig.Max
	if xConfig.Log && len(r.lineChart.ys) > 0 { // avoid log w/ empty data
		p.X.Scale = log2Scale{}
		p.X.Tick.Marker = log2Ticks{Prec: 1}
	}

	p.Y.Label.Text = yConfig.Label
	p.Y.Label.TextStyle.Color = theme.ForegroundColor()
	p.Y.LineStyle.Color = theme.ForegroundColor()
	p.Y.Tick.Label.Color = theme.ForegroundColor()
	p.Y.Tick.Color = theme.ForegroundColor()
	p.Y.Min = yConfig.Min
	p.Y.Max = yConfig.Max
	if yConfig.Log && len(r.lineChart.ys) > 0 { // avoid log w/ empty data
		p.Y.Scale = log2Scale{}
		p.Y.Tick.Marker = log2Ticks{Prec: 1}
	}

	r.lineChart.dataMtx.RLock()
	pts := make(plotter.XYs, len(r.lineChart.ys))
	for i := range r.lineChart.ys {
		var x, y float64
		if r.lineChart.xs == nil {
			x = float64(i)
		} else {
			x = r.lineChart.xs[i]
		}
		y = r.lineChart.ys[i]

		if xConfig.Clamp {
			x = math.Max(xConfig.Min, math.Min(xConfig.Max, x))
		} else {
			p.X.Min = math.Min(p.X.Min, x)
			p.X.Max = math.Max(p.X.Max, x)
		}

		if yConfig.Clamp {
			y = math.Max(yConfig.Min, math.Min(yConfig.Max, y))
		} else {
			p.Y.Min = math.Min(p.Y.Min, y)
			p.Y.Max = math.Max(p.Y.Max, y)
		}

		pts[i].X = x
		pts[i].Y = y
	}
	r.lineChart.dataMtx.RUnlock()

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

////////////////////////////////////////////////////////////////////////////////
// log2Scale

func (log2Scale) Normalize(min, max, x float64) float64 {
	if min <= 0 || max <= 0 || x <= 0 {
		panic("Values must be greater than 0 for a log2 scale.")
	}
	logMin := math.Log2(min)
	return (math.Log2(x) - logMin) / (math.Log2(max) - logMin)
}

////////////////////////////////////////////////////////////////////////////////
// log2Ticks

func (t log2Ticks) Ticks(min, max float64) []plot.Tick {
	if min <= 0 || max <= 0 {
		panic("Values must be greater than 0 for a log scale.")
	}

	val := math.Pow(2, float64(int(math.Log2(min))))
	max = math.Pow(2, float64(int(math.Ceil(math.Log2(max)))))
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
