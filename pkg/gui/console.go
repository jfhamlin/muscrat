package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

type (
	// Console is a fyne widget that displays the debug console.
	Console struct {
	}

	consoleEntry struct {
		text  string
		count int
	}
)

// NewConsole creates a new Console widget.
func NewConsole() fyne.CanvasObject {
	history := []consoleEntry{}

	historyText := func() string {
		sb := strings.Builder{}
		for _, entry := range history {
			if entry.count > 0 {
				sb.WriteString("(")
				sb.WriteString(fmt.Sprintf("%d", entry.count))
				sb.WriteString(") ")
			}

			sb.WriteString(entry.text)
		}
		return sb.String()
	}

	text := canvas.NewText(historyText(), theme.TextColor())

	pubsub.Subscribe("console.debug", func(event string, data any) {
		switch data := data.(type) {
		case string:
			if len(history) == 0 || history[len(history)-1].text != data {
				history = append(history, consoleEntry{text: data})
			} else {
				history[len(history)-1].count++
			}
			text.Text = historyText()
			text.Refresh()
		}
	})
	return container.New(nil, text)
}
