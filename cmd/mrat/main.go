package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/jfhamlin/muscrat/pkg/mrat"
)

func main() {
	flag.Parse()

	scriptFile := flag.Arg(0)
	if scriptFile == "" {
		fmt.Println("No script file provided")
		os.Exit(1)
	}

	script, err := ioutil.ReadFile(scriptFile)
	if err != nil {
		fmt.Printf("Error reading script file: %v\n", err)
		os.Exit(1)
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
			os.Exit(1)
		}
		close(done)
	}()

	// wait for OS interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	s := <-c
	fmt.Println("exiting on signal:", s)
	cancel()
	<-done
}
