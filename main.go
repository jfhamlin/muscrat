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
	// scriptFile := os.Args[1]
	// if scriptFile == "" {
	// 	fmt.Println("No script file provided")
	// 	os.Exit(1)
	// }

	// script, err := ioutil.ReadFile(scriptFile)
	// if err != nil {
	// 	fmt.Printf("Error reading script file: %v\n", err)
	// 	return
	// }

	msgs := make(chan *mrat.ServerMessage, 1)
	go func() {
		for msg := range msgs {
			fmt.Println("[SERVER]", msg.Text)
		}
	}()
	srv := mrat.NewServer(msgs)
	app := mrat.NewApp(srv)

	//srv.Start(string(script), scriptFile)

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// done := make(chan struct{})
	// go func() {
	// 	fmt.Println("Watching", scriptFile)
	// 	fmt.Println("Press Ctrl+C to exit")
	// 	err := mrat.WatchFile(ctx, scriptFile, srv)
	// 	if err != nil {
	// 		fmt.Printf("error watching file: %v\n", err)
	// 		return
	// 	}
	// 	close(done)
	// }()

	// Create application with options
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
