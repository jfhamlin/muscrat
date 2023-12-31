package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

type (
	// Console is a fyne widget that displays the debug console.
	Console struct {
	}

	consoleEntry struct {
		text   string
		count  int
		object fyne.CanvasObject
	}
)

// NewConsole creates a new Console widget.
func NewConsole() fyne.CanvasObject {
	history := []*consoleEntry{}
	entries := container.New(layout.NewVBoxLayout())

	pubsub.Subscribe("console.debug", func(event string, data any) {
		switch data := data.(type) {
		case string:
			if len(history) == 0 || history[len(history)-1].text != data {
				obj := canvas.NewText(data, theme.TextColor())
				history = append(history, &consoleEntry{
					text:   data,
					object: obj,
				})
				entries.AddObject(obj)
				entries.Refresh()
			} else {
				last := history[len(history)-1]
				last.inc()
			}
		}
	})
	return entries
}

func (ce *consoleEntry) String() string {
	sb := strings.Builder{}
	if ce.count > 0 {
		sb.WriteString("(")
		sb.WriteString(fmt.Sprintf("%d", ce.count))
		sb.WriteString(") ")
	}

	sb.WriteString(ce.text)
	return sb.String()
}

func (ce *consoleEntry) inc() {
	ce.count++
	ce.object.(*canvas.Text).Text = ce.String()
	ce.object.Refresh()
}
