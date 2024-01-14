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
		*fyne.Container

		history []*consoleEntry
		entries *fyne.Container
		title   fyne.CanvasObject
	}

	consoleEntry struct {
		text   string
		count  int
		object fyne.CanvasObject
	}
)

// NewConsole creates a new Console widget.
func NewConsole() fyne.CanvasObject {
	titleText := canvas.NewText("Console", theme.TextColor())
	title := container.New(layout.NewHBoxLayout(), titleText)

	console := &Console{
		history: []*consoleEntry{},
		title:   title,
		entries: container.New(layout.NewVBoxLayout()),
	}

	console.Container = container.New(layout.NewBorderLayout(title, nil, nil, nil), title, console.entries)

	pubsub.Subscribe("console.debug", console.onDebugMessage)
	return console.Container
}

func (c *Console) onDebugMessage(event string, data any) {
	switch data := data.(type) {
	case string:
		if len(c.history) == 0 || c.history[len(c.history)-1].text != data {
			obj := canvas.NewText(data, theme.TextColor())
			c.history = append(c.history, &consoleEntry{
				text:   data,
				object: obj,
			})
			c.entries.AddObject(obj)
			c.entries.Refresh()
		} else {
			last := c.history[len(c.history)-1]
			last.inc()
		}
	}
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
