package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/mrat"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

func main() {
	// ---- CLI flags ----------------------------------------------------------
	var (
		duration = flag.Int("duration", 10, "How long to run the script (seconds)")
	)
	flag.IntVar(duration, "d", 10, "Alias for -duration")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <script.glj>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	scriptFile := flag.Arg(0)

	// ---- Start muscrat DSP server ------------------------------------------
	srv := mrat.NewServer()
	srv.Start(context.Background(), true)

	// ---- Subscribe to sample stream & count samples ------------------------
	var totalSamples uint64 // counts *all* channel samples seen

	pubsub.Subscribe("samples", func(_ string, data any) {
		s, ok := data.([][]float64)
		if !ok {
			return
		}
		var n uint64
		ch := s[0]
		n += uint64(len(ch))
		atomic.AddUint64(&totalSamples, n)
	})

	// ---- Evaluate the script -----------------------------------------------
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := mrat.WatchScriptFile(ctx, scriptFile, srv); err != nil {
			log.Printf("script error: %v", err)
		}
	}()

	// ---- Run for the requested duration ------------------------------------
	time.Sleep(time.Duration(*duration) * time.Second)
	cancel() // stop the watcher / playback

	// ---- Print statistics ---------------------------------------------------
	elapsed := float64(*duration)
	samples := atomic.LoadUint64(&totalSamples)
	avgPerSec := samples / uint64(elapsed)

	fmt.Printf(
		"Ran %.0fs (%.0f Hz nominal)\n"+
			"Total samples  : %d\n"+
			"Average / sec  : %d\n",
		elapsed,
		float64(conf.SampleRate),
		samples,
		avgPerSec,
	)
}
