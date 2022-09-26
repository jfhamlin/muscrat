package graph

import (
	"context"
	"sync"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
)

type NodeID int

// Node is a node in a graph of SampleGenerators.
type Node struct {
	ID        NodeID
	Generator generator.SampleGenerator
}

type Edge struct {
	From    NodeID
	To      NodeID
	Channel chan []float64
}

// Graph is a graph of SampleGenerators.
type Graph struct {
	Nodes []*Node
	Edges []*Edge
}

func (g *Graph) IncomingEdges(id NodeID) []*Edge {
	var edges []*Edge
	for _, e := range g.Edges {
		if e.To == id {
			edges = append(edges, e)
		}
	}
	return edges
}

func (g *Graph) OutgoingEdges(id NodeID) []*Edge {
	var edges []*Edge
	for _, e := range g.Edges {
		if e.From == id {
			edges = append(edges, e)
		}
	}
	return edges
}

func RunGraph(ctx context.Context, g *Graph, cfg generator.SampleConfig) {
	var wg sync.WaitGroup
	for _, node := range g.Nodes {
		wg.Add(1)
		go func(n *Node) {
			RunNode(ctx, n, g, cfg)
			wg.Done()
		}(node)
	}
	wg.Wait()
}

func RunNode(ctx context.Context, node *Node, g *Graph, cfg generator.SampleConfig) {
	inEdges := g.IncomingEdges(node.ID)
	inputSamples := make([]float64, len(inEdges))
	for i, e := range inEdges {
		inputSamples[i] = <-e.Channel
	}

	outputSamples := node.Generator.GenerateSamples(ctx, cfg, n)
	outEdges := g.IncomingEdges(node.ID)
	for _, e := range outEdges {
		e.Channel <- outputSamples
	}
}
