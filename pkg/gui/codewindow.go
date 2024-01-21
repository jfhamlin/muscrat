package gui

import (
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

type (
	CodeWindow struct {
		fyne.Window

		fileName string
	}
)

func NewCodeWindow(a fyne.App, fileName string) *CodeWindow {
	w := a.NewWindow("muscrat editor")

	absPath, err := filepath.Abs(fileName)
	if err != nil {
		absPath = fileName // if we can't get the absolute path, just use the original
	}

	fileContents, err := os.ReadFile(fileName)
	if err != nil {
		pubsub.Publish("console.debug", fmt.Sprintf("error reading file %s: %v", fileName, err))
	}

	// toolbar has a new file button, a load button, a save button, and the name of the
	// file being edited. below that is a tabbed pane. the first tab is
	// the editor. the second tab is the console.

	newButton := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
		fmt.Println("new")
	})

	loadButton := widget.NewButtonWithIcon("", theme.FolderIcon(), func() {
		fmt.Println("load")
	})

	saveButton := widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		fmt.Println("save")
	})

	toolbar := container.NewHBox(newButton, loadButton, saveButton, canvas.NewText(absPath, theme.TextColor()))

	editor := widget.NewMultiLineEntry()
	editor.SetText(string(fileContents))

	contents := container.New(layout.NewBorderLayout(toolbar, nil, nil, nil), toolbar, editor)

	w.SetContent(contents)

	return &CodeWindow{
		Window:   w,
		fileName: fileName,
	}
}
