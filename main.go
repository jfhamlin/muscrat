package main

import (
	"embed"
	"fmt"

	"github.com/jfhamlin/muscrat/pkg/mrat"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	msgs := make(chan *mrat.ServerMessage, 1)
	go func() {
		for msg := range msgs {
			fmt.Println("[SERVER]", msg.Text)
		}
	}()
	srv := mrat.NewServer(msgs)
	app := mrat.NewApp(srv)

	err := wails.Run(&options.App{
		Title:            "muscrat",
		Width:            1024,
		Height:           768,
		Assets:           assets,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			srv,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
