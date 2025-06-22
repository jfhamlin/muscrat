package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/mrat"
	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

func main() {
	// ---------------------------------------------------------------- CLI ----
	var durSec = flag.Int("duration", 10, "How long to run the script (seconds)")
	flag.IntVar(durSec, "d", 10, "Alias for -duration")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <script.glj>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	scriptFile := flag.Arg(0)

	// -------------------------------------------------------- start server ----
	srv := mrat.NewServer()
	srv.Start(context.Background(), false)

	// --------------------------------------------------- counters & latency ---
	var (
		totalSamples uint64 // first-channel samples only
		lastTS       time.Time
		latencies    []float64 // seconds between buffers
		latMtx       sync.Mutex
	)

	pubsub.Subscribe("samples", func(_ string, data any) {
		s, ok := data.([][]float64)
		if !ok || len(s) == 0 {
			return
		}

		// 1) sample count (first channel only)
		atomic.AddUint64(&totalSamples, uint64(len(s[0])))

		// 2) latency stats
		now := time.Now()
		latMtx.Lock()
		if !lastTS.IsZero() {
			latencies = append(latencies, now.Sub(lastTS).Seconds())
		}
		lastTS = now
		latMtx.Unlock()
	})

	// ---------------------------------------------- evaluate + run duration ---
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := mrat.WatchScriptFile(ctx, scriptFile, srv); err != nil {
			log.Printf("script error: %v", err)
		}
	}()

	time.Sleep(time.Duration(*durSec) * time.Second)
	cancel()

	// ----------------------------------------------------- sample statistics --
	elapsed := float64(*durSec)
	samples := atomic.LoadUint64(&totalSamples)
	fmt.Printf("Ran %.0fs (%.0f Hz nominal)\n", elapsed, float64(conf.SampleRate))
	fmt.Printf("Total samples (ch-0): %d  |  Average/s: %d\n\n",
		samples, samples/uint64(elapsed))

	// ----------------------------------------------------- latency statistics --
	latMtx.Lock()
	lats := append([]float64(nil), latencies...) // copy under lock
	latMtx.Unlock()

	if len(lats) == 0 {
		fmt.Println("No latency data recorded.")
		return
	}

	sort.Float64s(lats)
	sum := 0.0
	for _, v := range lats {
		sum += v
	}
	n := float64(len(lats))
	quantile := func(p float64) float64 {
		if n == 1 {
			return lats[0]
		}
		idx := p * (n - 1)
		lo := int(math.Floor(idx))
		hi := int(math.Ceil(idx))
		if lo == hi {
			return lats[lo]
		}
		frac := idx - float64(lo)
		return lats[lo]*(1-frac) + lats[hi]*frac
	}

	nominalLatency := float64(conf.BufferSize) / float64(conf.SampleRate)
	fmt.Println("Latency between successive buffers (ms):")
	fmt.Printf("  nominal: %.2f\n", nominalLatency*1000)
	fmt.Printf("  min : %.2f\n", lats[0]*1000)
	fmt.Printf("  p25 : %.2f\n", quantile(0.25)*1000)
	fmt.Printf("  p50 : %.2f\n", quantile(0.50)*1000)
	fmt.Printf("  p75 : %.2f\n", quantile(0.75)*1000)
	fmt.Printf("  max : %.2f\n", lats[len(lats)-1]*1000)
	fmt.Printf("  mean: %.2f\n", sum/n*1000)
}
