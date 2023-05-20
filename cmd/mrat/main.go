package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/jfhamlin/muscrat/pkg/aio"
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
	{
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	script, err := ioutil.ReadFile(scriptFile)
	if err != nil {
		fmt.Printf("Error reading script file: %v\n", err)
		return
	}

	msgs := make(chan *mrat.ServerMessage, 1)
	go func() {
		for msg := range msgs {
			fmt.Println("[SERVER]", msg.Text)
		}
	}()

	srv := mrat.NewServer(msgs)
	srv.Start(string(script), scriptFile)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		fmt.Println("Watching", scriptFile)
		fmt.Println("Press Ctrl+C to exit")
		err := mrat.WatchFile(ctx, scriptFile, srv)
		if err != nil {
			fmt.Printf("error watching file: %v\n", err)
			return
		}
		close(done)
	}()

	for {
		b := make([]byte, 1)
		_, err = os.Stdin.Read(b)
		if err != nil {
			fmt.Println(err)
			return
		}
		if b[0] == 3 {
			fmt.Println("exiting")
			cancel()
			return
		}
		select {
		case aio.StdinChan <- b[0]:
		default:
		}
	}
}
