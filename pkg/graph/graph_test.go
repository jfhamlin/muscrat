package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

// goos: darwin
// goarch: arm64
// pkg: github.com/jfhamlin/muscrat/pkg/graph
// BenchmarkGraph-8   	  180675	      6712 ns/op
// PASS
// ok  	github.com/jfhamlin/muscrat/pkg/graph	2.360s
func BenchmarkGraph(b *testing.B) {
	//	timings := make([]time.Duration, b.N)

	g := benchGraph()
	sinks := g.OutputChans()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go g.Run(ctx, ugen.SampleConfig{SampleRateHz: 44100})
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

	// // add disconnected nodes to the graph to test for races
	// {
	// 	pn := g.AddGeneratorNode(ugen.NewProduct())
	// 	for i := 0; i < 20; i++ {
	// 		cn := g.AddGeneratorNode(ugen.NewConstant(1.0))
	// 		g.AddEdge(cn.ID(), pn.ID(), fmt.Sprintf("constant-%d", i))
	// 	}
	// }

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

	sn := g.AddOutNode()
	g.AddEdge(sum.ID(), sn.ID(), "in")

	return g
}

func TestCycle(t *testing.T) {
	g := &Graph{
		BufferSize: 128,
	}

	n1 := g.AddGeneratorNode(ugen.NewConstant(1.0))
	n2 := g.AddGeneratorNode(ugen.NewSum())
	n3 := g.AddGeneratorNode(ugen.NewSum())

	g.AddEdge(n1.ID(), n2.ID(), "n1->n2")
	g.AddEdge(n2.ID(), n3.ID(), "n2->n3")
	g.AddEdge(n3.ID(), n2.ID(), "n3->n2")

	out := g.AddOutNode()
	g.AddEdge(n3.ID(), out.ID(), "n3->out")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go g.Run(ctx, ugen.SampleConfig{SampleRateHz: 44100})

	var result []float64
	for i := 0; i < 10; i++ {
		select {
		case res, ok := <-out.Chan():
			if !ok {
				t.Fail()
			}
			result = append(result, res[0])
		case <-time.After(10 * time.Second):
			t.Fatal("timeout")
		}
	}
	// it should = [1, 2, ..., 10]
	for i := 0; i < 10; i++ {
		if result[i] != float64(i+1) {
			t.Errorf("expected %d, got %f", i+1, result[i])
		}
	}
}

func FuzzGraphLiveness(f *testing.F) {
	type edge struct {
		From int `json:"from"`
		To   int `json:"to"`
	}

	type testGraph struct {
		NodeCount int    `json:"nodeCount"`
		Edges     []edge `json:"edges"`
		OutNode   int    `json:"outNode"`
	}

	seeds := []string{
		`{
       "nodeCount": 2,
       "edges": [{"from": 0, "to": 1}],
       "outNode": 1
     }`,
		`{
       "nodeCount": 4,
       "edges": [{"from": 0, "to": 1}, {"from": 1, "to": 2}, {"from": 2, "to": 1}, {"from": 2, "to": 3}],
       "outNode": 3
     }`,
	}

	for _, seed := range seeds {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, b []byte) {
		var tg testGraph
		if err := json.Unmarshal(b, &tg); err != nil {
			t.Skip()
		}
		if len(tg.Edges) == 0 || tg.NodeCount <= 0 {
			t.Skip()
		}

		// normalize node ids
		tg.OutNode = tg.OutNode % tg.NodeCount
		for i := range tg.Edges {
			edge := &tg.Edges[i]
			edge.From = edge.From % tg.NodeCount
			edge.To = edge.To % tg.NodeCount
		}

		g := &Graph{
			BufferSize: 128,
		}

		var nodes []Node
		var outNode *OutNode
		for i := 0; i < tg.NodeCount; i++ {
			var node Node
			if i == tg.OutNode {
				outNode = g.AddOutNode()
				node = outNode
			} else {
				node = g.AddGeneratorNode(ugen.NewConstant(1.0))
			}
			nodes = append(nodes, node)
		}

		for _, edge := range tg.Edges {
			fromIndex := edge.From
			if fromIndex == tg.OutNode {
				t.Skip()
			}

			from := nodes[fromIndex].ID()
			to := nodes[edge.To].ID()
			if from == to {
				t.Skip()
			}
			g.AddEdge(from, to, fmt.Sprintf("%v -> %v", from, to))
		}

		if len(g.IncomingEdges(outNode.ID())) == 0 {
			// TODO: implement graph validation
			t.Skip()
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go g.Run(ctx, ugen.SampleConfig{SampleRateHz: 44100})

		for i := 0; i < 10; i++ {
			select {
			case _, ok := <-outNode.Chan():
				if !ok {
					t.Fail()
				}
			case <-time.After(1 * time.Second):
				t.Fatal("timeout")
			}
		}
	})
}
