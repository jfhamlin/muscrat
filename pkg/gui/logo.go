package gui

import (
	"embed"

	"fyne.io/fyne/v2/canvas"
)

const (
	logoImagePath = "assets/images/muscrat.svg"
)

var (
	//go:embed assets/*
	assetsFS embed.FS
)

func LogoImage() *canvas.Image {
	f, err := assetsFS.Open(logoImagePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	logo := canvas.NewImageFromReader(f, logoImagePath)
	logo.FillMode = canvas.ImageFillContain
	return logo
}
