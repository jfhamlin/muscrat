package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// CollapsibleSidebar defines a container with a collapsible sidebar.
type CollapsibleSidebar struct {
	widget.BaseWidget
	Content     fyne.CanvasObject
	Collapsed   bool
	LeftSidebar bool
	button      *widget.Button
	spacer      *canvas.Rectangle
	spacerWidth float32
}

// NewCollapsibleSidebar creates a new collapsible sidebar with the specified content.
func NewCollapsibleSidebar(content fyne.CanvasObject, leftSidebar bool) *CollapsibleSidebar {
	sidebar := &CollapsibleSidebar{
		Content:     content,
		LeftSidebar: leftSidebar,
		Collapsed:   false,
		spacerWidth: theme.Padding(),
	}

	sidebar.button = widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), func() {
		sidebar.Collapsed = !sidebar.Collapsed
		sidebar.Refresh()
	})
	sidebar.spacer = canvas.NewRectangle(theme.ShadowColor())

	sidebar.ExtendBaseWidget(sidebar)
	return sidebar
}

// CreateRenderer links this widget to its renderer.
func (s *CollapsibleSidebar) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)
	return &collapsibleSidebarRenderer{
		sidebar: s,
		objects: []fyne.CanvasObject{s.button, s.spacer, s.Content},
	}
}

type collapsibleSidebarRenderer struct {
	sidebar *CollapsibleSidebar
	objects []fyne.CanvasObject
}

func (r *collapsibleSidebarRenderer) Layout(size fyne.Size) {
	buttonSize := r.sidebar.button.MinSize()
	spacerWidth := r.sidebar.spacerWidth

	if r.sidebar.Collapsed {
		if !r.sidebar.LeftSidebar {
			r.sidebar.button.Move(fyne.NewPos(0, (size.Height-buttonSize.Height)/2))
		} else {
			r.sidebar.button.Move(fyne.NewPos(size.Width-buttonSize.Width, (size.Height-buttonSize.Height)/2))
		}
		r.sidebar.button.Resize(buttonSize)

		r.sidebar.Content.Hide()
		r.sidebar.spacer.Hide()
	} else {
		if !r.sidebar.LeftSidebar {
			r.sidebar.spacer.Move(fyne.NewPos(0, 0))
			r.sidebar.button.Move(fyne.NewPos(spacerWidth, (size.Height-buttonSize.Height)/2))
			r.sidebar.Content.Move(fyne.NewPos(spacerWidth+buttonSize.Width, 0))
		} else {
			r.sidebar.spacer.Move(fyne.NewPos(size.Width-spacerWidth, 0))
			r.sidebar.button.Move(fyne.NewPos(size.Width-spacerWidth-buttonSize.Width, (size.Height-buttonSize.Height)/2))
			r.sidebar.Content.Move(fyne.NewPos(0, 0))
		}
		r.sidebar.spacer.Resize(fyne.NewSize(spacerWidth, size.Height))
		r.sidebar.button.Resize(buttonSize)
		r.sidebar.Content.Resize(fyne.NewSize(size.Width-spacerWidth-buttonSize.Width, size.Height))

		r.sidebar.Content.Show()
		r.sidebar.spacer.Show()
	}
}

func (r *collapsibleSidebarRenderer) MinSize() fyne.Size {
	if r.sidebar.Collapsed {
		return r.sidebar.button.MinSize()
	}
	return fyne.NewSize(
		r.sidebar.Content.MinSize().Width+r.sidebar.spacerWidth,
		r.sidebar.Content.MinSize().Height)
}

func (r *collapsibleSidebarRenderer) Refresh() {
	r.Layout(r.sidebar.Size())
	canvas.Refresh(r.sidebar)
}

func (r *collapsibleSidebarRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *collapsibleSidebarRenderer) Destroy() {}
