package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func BenchmarkGraph(b *testing.B) {
	g := benchGraph()
	sinks := g.SinkChans()
	go g.Run(context.Background(), ugen.SampleConfig{SampleRateHz: 44100})
	for n := 0; n < b.N; n++ {
		for _, sink := range sinks {
			<-sink
		}
	}
}

func benchGraph() *Graph {
	g := &Graph{
		BufferSize: 128,
	}

	var constants []Node
	for i := 0; i < 100; i++ {
		n := g.AddGeneratorNode(ugen.NewConstant(1.0))
		constants = append(constants, n)
	}

	var products []Node
	for i := 0; i < 100; i++ {
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
