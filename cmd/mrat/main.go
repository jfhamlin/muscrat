package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"fyne.io/fyne/v2/app"

	"github.com/jfhamlin/muscrat/pkg/gui"
	"github.com/jfhamlin/muscrat/pkg/mrat"

	"golang.org/x/term"

	// pprof
	"net/http"
	_ "net/http/pprof"
)

func init() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func main() {
	flag.Parse()

	scriptFile := flag.Arg(0)
	if scriptFile == "" {
		fmt.Println("No script file provided")
		os.Exit(1)
	}

	// set up raw input mode
	if false {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	msgs := make(chan *mrat.ServerMessage, 1)
	go func() {
		for msg := range msgs {
			fmt.Println("[SERVER]", msg.Text)
		}
	}()

	srv := mrat.NewServer(msgs)
	srv.Start(context.Background())

	if err := srv.EvalScript(scriptFile); err != nil {
		fmt.Printf("error evaluating script: %v\n", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	// wait for kill signal
	done := make(chan os.Signal, 1)
	go signal.Notify(done, os.Interrupt)
	go func() {
		<-done
		cancel()
	}()

	go func() {
		if err := mrat.WatchScriptFile(ctx, scriptFile, srv); err != nil {
			fmt.Printf("error watching script file: %v\n", err)
			return
		}
		os.Exit(0)
	}()

	{
		a := app.New()
		main := gui.NewMainWindow(a)
		main.SetMaster()

		//code := gui.NewCodeWindow(a, scriptFile)

		main.Show()
		//code.Show()

		a.Run()
	}
}
