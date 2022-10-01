package graph

import (
	"context"
	"sync"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
)

type NodeID int

type Node interface {
	ID() NodeID
	Run(ctx context.Context, g *Graph, cfg generator.SampleConfig, numSamples int)
}

// Node is a node in a graph of nodes.
type GeneratorNode struct {
	id        NodeID
	Generator generator.SampleGenerator
}

func (n *GeneratorNode) ID() NodeID {
	return n.id
}

func (n *GeneratorNode) Run(ctx context.Context, g *Graph, cfg generator.SampleConfig, numSamples int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		RunNode(ctx, n, g, cfg, numSamples)
	}
}

type SinkNode struct {
	id     NodeID
	Output chan []float64
}

func (n *SinkNode) ID() NodeID {
	return n.id
}

func (n *SinkNode) Run(ctx context.Context, g *Graph, cfg generator.SampleConfig, numSamples int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		inEdges := g.IncomingEdges(n.id)
		inputSamples := make([][]float64, len(inEdges))
		for i, e := range inEdges {
			inputSamples[i] = <-e.Channel
		}
		if len(inputSamples) == 0 {
			continue
		}
		n.Output <- inputSamples[0]
	}
}

type Edge struct {
	From    NodeID
	To      NodeID
	Channel chan []float64
}

// Graph is a graph of SampleGenerators.
type Graph struct {
	Nodes []Node
	Edges []*Edge
}

func (g *Graph) AddEdge(from, to NodeID) {
	g.Edges = append(g.Edges, &Edge{
		From:    from,
		To:      to,
		Channel: make(chan []float64),
	})
}

func (g *Graph) AddGeneratorNode(gen generator.SampleGenerator) NodeID {
	node := &GeneratorNode{
		id:        NodeID(len(g.Nodes)),
		Generator: gen,
	}
	g.Nodes = append(g.Nodes, node)
	return node.id
}

func (g *Graph) AddSinkNode() (NodeID, <-chan []float64) {
	node := &SinkNode{
		id:     NodeID(len(g.Nodes)),
		Output: make(chan []float64),
	}
	g.Nodes = append(g.Nodes, node)
	return node.id, node.Output
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
		go func(n Node) {
			n.Run(ctx, g, cfg, 1024)
			wg.Done()
		}(node)
	}
	wg.Wait()
}

func RunNode(ctx context.Context, node *GeneratorNode, g *Graph, cfg generator.SampleConfig, numSamples int) {
	inEdges := g.IncomingEdges(node.ID())
	inputSamples := make([][]float64, len(inEdges))
	for i, e := range inEdges {
		inputSamples[i] = <-e.Channel
	}

	cfg.InputSamples = inputSamples
	outputSamples := node.Generator.GenerateSamples(ctx, cfg, numSamples)
	outEdges := g.OutgoingEdges(node.ID())
	for _, e := range outEdges {
		e.Channel <- outputSamples
	}
}
