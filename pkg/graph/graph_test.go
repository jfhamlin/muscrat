package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

////////////////////////////////////////////////////////////////////////////////
// CONTROL BENCHMARK (goroutines and channels)
// goos: darwin
// goarch: arm64
// pkg: github.com/jfhamlin/muscrat/pkg/graph
// BenchmarkGraph-8   	   26818	     51012 ns/op
// PASS
// ok  	github.com/jfhamlin/muscrat/pkg/graph	1.927s

func BenchmarkGraph(b *testing.B) {
	//timings := make([]time.Duration, b.N)

	g := benchGraph()
	sinks := g.SinkChans()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go g.Run(ctx, ugen.SampleConfig{SampleRateHz: 44100})
	for n := 0; n < b.N; n++ {
		//start := time.Now()
		for _, sink := range sinks {
			<-sink
		}
		//timings[n] = time.Since(start)
	}

	// // print average time and p90 time
	// sort.Slice(timings, func(i, j int) bool {
	// 	return timings[i] < timings[j]
	// })
	// var total time.Duration
	// for _, timing := range timings {
	// 	total += timing
	// }
	// fmt.Println("---", b.N)
	// fmt.Println("old p50", timings[int(float64(b.N)*0.5)])
	// fmt.Println("old p90", timings[int(float64(b.N)*0.9)])
	// fmt.Println("old avg", total/time.Duration(b.N))
}

func BenchmarkGraphWorkers(b *testing.B) {
	//	timings := make([]time.Duration, b.N)

	g := benchGraph()
	sinks := g.SinkChans()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go g.RunWorkers(ctx, ugen.SampleConfig{SampleRateHz: 44100})
	for n := 0; n < b.N; n++ {
		//		start := time.Now()
		for _, sink := range sinks {
			<-sink
		}
		//timings[n] = time.Since(start)
	}

	// // print average time and p90 time
	// sort.Slice(timings, func(i, j int) bool {
	// 	return timings[i] < timings[j]
	// })
	// var total time.Duration
	// for _, timing := range timings {
	// 	total += timing
	// }
	// fmt.Println("---", b.N)
	// fmt.Println("new p50", timings[int(float64(b.N)*0.5)])
	// fmt.Println("new p90", timings[int(float64(b.N)*0.9)])
	// fmt.Println("new avg", total/time.Duration(b.N))
}

func benchGraph() *Graph {
	g := &Graph{
		BufferSize: 128,
	}

	// add disconnected nodes to the graph to test for races
	{
		pn := g.AddGeneratorNode(ugen.NewProduct())
		for i := 0; i < 20; i++ {
			cn := g.AddGeneratorNode(ugen.NewConstant(1.0))
			g.AddEdge(cn.ID(), pn.ID(), fmt.Sprintf("constant-%d", i))
		}
	}

	const (
		numConsts   = 10
		numProducts = 10
	)

	var constants []Node
	for i := 0; i < numConsts; i++ {
		n := g.AddGeneratorNode(ugen.NewConstant(1.0))
		constants = append(constants, n)
	}

	var products []Node
	for i := 0; i < numProducts; i++ {
		n := g.AddGeneratorNode(ugen.NewProduct())
		products = append(products, n)

		for i, constant := range constants {
			g.AddEdge(constant.ID(), n.ID(), fmt.Sprintf("constant-%d", i))
		}
	}

	sum := g.AddGeneratorNode(ugen.NewSum())
	for i, product := range products {
		g.AddEdge(product.ID(), sum.ID(), fmt.Sprintf("product-%d", i))
	}

	sn := g.AddSinkNode()
	g.AddEdge(sum.ID(), sn.ID(), "in")

	return g
}
