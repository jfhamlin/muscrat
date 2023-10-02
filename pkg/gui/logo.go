package gui

import (
	"embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

const (
	logoImagePath     = "assets/images/muscrat.svg"
	logoImagePathDark = "assets/images/muscrat-dark.svg"
)

var (
	//go:embed assets/*
	assetsFS embed.FS
)

func LogoImage() *canvas.Image {
	path := logoImagePath
	if fyne.CurrentApp().Settings().ThemeVariant() == theme.VariantDark {
		path = logoImagePathDark
	}

	f, err := assetsFS.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	logo := canvas.NewImageFromReader(f, path)
	logo.FillMode = canvas.ImageFillContain
	return logo
}
