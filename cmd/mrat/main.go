package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

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
	if false {
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

	_ = script

	msgs := make(chan *mrat.ServerMessage, 1)
	go func() {
		for msg := range msgs {
			fmt.Println("[SERVER]", msg.Text)
		}
	}()

	srv := mrat.NewServer(msgs)
	srv.Start(context.Background())

	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx

	done := make(chan struct{})
	go func() {
		fmt.Println("Watching", scriptFile)
		fmt.Println("Press Ctrl+C to exit")
		// err := mrat.WatchFile(ctx, scriptFile, srv)
		// if err != nil {
		// 	fmt.Printf("error watching file: %v\n", err)
		// 	return
		// }
		close(done)
	}()

	streamKeys(cancel)
	<-done
}

func streamKeys(cancel context.CancelFunc) {
	defer cancel()

	erred := make(chan error)
	keyCh := make(chan byte, 1)
	go func() {
		for {
			b := make([]byte, 1)
			_, err := os.Stdin.Read(b)
			if err != nil {
				erred <- err
				return
			}
			keyCh <- b[0]
		}
	}()

	// track one key input at a time for now
	var pressedKey byte
	const keyTimeout = 500 * time.Millisecond
	for {
		var b byte
		select {
		case b = <-keyCh:
		case <-time.After(keyTimeout):
			if pressedKey != 0 {
				pressedKey = 0
			}
		case err := <-erred:
			fmt.Println("Error reading from stdin:", err)
			return
		}
		if b == 3 {
			fmt.Println("exiting")
			return
		}
		pressedKey = b
		select {
		case aio.StdinChan <- b:
		default:
		}
	}
}
