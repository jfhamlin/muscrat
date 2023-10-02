package meter

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	numTicks  = 11
	peakWidth = 4
)

type (
	// Volume is a struct that represents a volume meter.
	Volume struct {
		widget.BaseWidget

		Min, Max, Value float64
		Peak            float64

		minSize fyne.Size
	}

	volumeRenderer struct {
		volume *Volume

		background, greenBar, yellowBar, redBar, peakBar *canvas.Rectangle

		volumeLabel *canvas.Text
		tickLabels  []*canvas.Text

		objects []fyne.CanvasObject
	}
)

var (
	_ fyne.WidgetRenderer = (*volumeRenderer)(nil)
)

// NewVolume creates a new Volume meter.
func NewVolume(min, max float64) *Volume {
	v := &Volume{
		Min: min,
		Max: max,
	}
	v.ExtendBaseWidget(v)

	return v
}

func (v *Volume) SetMinSize(size fyne.Size) {
	v.minSize = size
}

func (v *Volume) SetValues(value, peak float64) {
	v.Value = value
	v.Peak = peak
	v.Refresh()
}

func (v *Volume) CreateRenderer() fyne.WidgetRenderer {
	background := canvas.NewRectangle(theme.BackgroundColor())

	redBar := canvas.NewRectangle(color.RGBA{R: 255, A: 255})
	yellowBar := canvas.NewRectangle(color.RGBA{R: 255, G: 255, A: 255})
	greenBar := canvas.NewRectangle(color.RGBA{G: 255, A: 255})

	peakBar := canvas.NewRectangle(theme.PrimaryColor())

	volumeLabel := canvas.NewText("0dB", theme.BackgroundColor())
	tickLabels := []*canvas.Text{}
	for i := 0; i < numTicks; i++ {
		label := canvas.NewText("", theme.TextColor())
		// monospace font
		label.TextStyle.Monospace = true
		tickLabels = append(tickLabels, label)
	}

	objects := []fyne.CanvasObject{redBar, yellowBar, greenBar, peakBar, background, volumeLabel}
	for i := range tickLabels {
		objects = append(objects, tickLabels[i])
	}

	return &volumeRenderer{
		volume:      v,
		background:  background,
		redBar:      redBar,
		yellowBar:   yellowBar,
		greenBar:    greenBar,
		peakBar:     peakBar,
		volumeLabel: volumeLabel,
		tickLabels:  tickLabels,
		objects:     objects,
	}
}

////////////////////////////////////////////////////////////////////////////////
// Renderer

func (r *volumeRenderer) update() {
	value := r.volume.Value
	if r.volume.Value < r.volume.Min {
		value = r.volume.Min
	}
	if r.volume.Value > r.volume.Max {
		value = r.volume.Max
	}
	// place tick labels along the left side of the bar, centered
	// around the tick value's placement on the bar.
	// the largest label is the max value, placed at the top
	// the smallest label is the min value, placed at the bottom
	// the bars take up the rest of the space to the right, with some
	// padding.

	// the bars are placed on top of each other, with the red bar
	// on top, then the yellow bar, then the green bar.
	// the green bar is up to -10dB, the yellow bar is up to -3dB,
	// and the red bar is up to 0dB.
	// the yellow and red bars are hidden if the volume is below
	// their threshold.
	// the green bar is always visible, but shrinks to fit the
	// volume.

	greenThreshold := -10.0
	yellowThreshold := -3.0

	delta := r.volume.Max - r.volume.Min
	ratio := (value - r.volume.Min) / delta
	maxGreenRatio := (greenThreshold - r.volume.Min) / delta
	maxYellowRatio := (yellowThreshold - r.volume.Min) / delta

	size := r.volume.Size()

	const labelPadding = 2
	maxLabelWidth := float32(0)
	for i := 0; i < numTicks; i++ {
		label := r.tickLabels[i]
		tickValue := r.volume.Max - delta*float64(i)/(numTicks-1)
		label.Text = fmt.Sprintf("%.0fdB", tickValue)
		if i == 0 {
			label.Move(fyne.NewPos(0, 0))
		} else if i == numTicks-1 {
			label.Move(fyne.NewPos(0, size.Height-label.MinSize().Height))
		} else {
			label.Move(fyne.NewPos(0, size.Height*float32(i)/(numTicks-1)-label.MinSize().Height/2))
		}
		if label.MinSize().Width > maxLabelWidth {
			maxLabelWidth = label.MinSize().Width
		}
	}
	meterLeft := maxLabelWidth + labelPadding
	meterSize := fyne.NewSize(size.Width-meterLeft, size.Height)

	r.background.Resize(meterSize)
	r.background.Move(fyne.NewPos(meterLeft, 0))

	{ // bars
		r.greenBar.Resize(fyne.NewSize(meterSize.Width, meterSize.Height*float32(math.Min(ratio, maxGreenRatio))))
		r.greenBar.Move(fyne.NewPos(meterLeft, meterSize.Height-r.greenBar.Size().Height))

		if value <= greenThreshold {
			r.redBar.Hide()
			r.yellowBar.Hide()
		} else {
			r.yellowBar.Show()
			r.yellowBar.Resize(fyne.NewSize(meterSize.Width, meterSize.Height*float32(math.Min(ratio, maxYellowRatio))))
			r.yellowBar.Move(fyne.NewPos(meterLeft, meterSize.Height-r.yellowBar.Size().Height))
			if value <= yellowThreshold {
				r.redBar.Hide()
			} else {
				r.redBar.Show()
				r.redBar.Resize(fyne.NewSize(meterSize.Width, meterSize.Height*float32(ratio)))
				r.redBar.Move(fyne.NewPos(meterLeft, meterSize.Height-r.redBar.Size().Height))
			}
		}
		// peak bar is on the right side of the meter, and is peakWidth wide
		// its height is the value of the peak
		peakColor := theme.PrimaryColor()
		if r.volume.Peak > 1 {
			r.volume.Peak = 1
			peakColor = theme.ErrorColor()
		}
		if r.volume.Peak < 0 {
			r.volume.Peak = 0
		}
		r.peakBar.FillColor = peakColor
		r.peakBar.Resize(fyne.NewSize(peakWidth, meterSize.Height*float32(r.volume.Peak)))
		r.peakBar.Move(fyne.NewPos(meterLeft+meterSize.Width-peakWidth, meterSize.Height-r.peakBar.Size().Height))
	}

	r.volumeLabel.Text = fmt.Sprintf("%.0fdB", r.volume.Value)
	// place the label horizontally centered and vertically just below
	// the top of the bar
	r.volumeLabel.Move(fyne.NewPos(
		meterLeft+(meterSize.Width-r.volumeLabel.MinSize().Width)/2,
		float32(math.Min(float64(meterSize.Height)*(1-ratio), float64(meterSize.Height-r.volumeLabel.MinSize().Height))),
	))
}

func (r *volumeRenderer) applyTheme() {
	r.background.FillColor = color.Transparent
	r.background.StrokeColor = theme.TextColor()
	r.background.StrokeWidth = 1

	r.peakBar.FillColor = theme.PrimaryColor()

	r.volumeLabel.TextSize = theme.TextSize()
	r.volumeLabel.Color = theme.PrimaryColor()

	textColor := theme.TextColor()
	textSize := 2 * theme.TextSize() / 3
	for _, label := range r.tickLabels {
		label.Color = textColor
		label.TextSize = textSize
	}
}

func (r *volumeRenderer) Layout(size fyne.Size) {
	r.update()
}

func (r *volumeRenderer) MinSize() fyne.Size {
	return r.volume.minSize
}

func (r *volumeRenderer) Refresh() {
	r.applyTheme()
	r.update()
	r.background.Refresh()
	r.redBar.Refresh()
	r.yellowBar.Refresh()
	r.greenBar.Refresh()
	r.volumeLabel.Refresh()
	canvas.Refresh(r.volume)
}

func (r *volumeRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *volumeRenderer) Destroy() {}
