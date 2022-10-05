package graph

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
)

type NodeID int

func (id NodeID) String() string {
	return strconv.Itoa(int(id))
}

type Node interface {
	ID() NodeID
	Run(ctx context.Context, g *Graph, cfg generator.SampleConfig, numSamples int)

	String() string
}

type nodeOptions struct {
	label string
}

// NodeOptions is a functional option for configuring a node.
type NodeOption func(*nodeOptions)

func WithLabel(label string) NodeOption {
	return func(o *nodeOptions) {
		o.label = label
	}
}

// Node is a node in a graph of nodes.
type GeneratorNode struct {
	id        NodeID
	Generator generator.SampleGenerator
	label     string
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

func (n *GeneratorNode) GenerateSamples(ctx context.Context, cfg generator.SampleConfig, numSamples int) (outputSamples []float64) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: make the failure of this node visible to the user.
			//fmt.Printf("node %s failed: %v\n", n, r)
			// print stack trace
			fmt.Printf("stack trace: %s\n", debug.Stack())
			outputSamples = make([]float64, numSamples)
		}
	}()
	return n.Generator.GenerateSamples(ctx, cfg, numSamples)
}

func (n *GeneratorNode) String() string {
	if n.label != "" {
		return n.label
	}
	return n.ID().String()
}

type SinkNode struct {
	id     NodeID
	Output chan []float64
	label  string
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
			select {
			case <-ctx.Done():
				return
			case inputSamples[i] = <-e.Channel:
			}
		}
		if len(inputSamples) == 0 {
			continue
		}
		select {
		case n.Output <- inputSamples[0]:
		case <-ctx.Done():
			return
		}
	}
}

func (n *SinkNode) String() string {
	if n.label != "" {
		return n.label
	}
	return n.ID().String()
}

type Edge struct {
	From    NodeID
	To      NodeID
	ToPort  string
	Channel chan []float64
}

// Graph is a graph of SampleGenerators.
type Graph struct {
	Nodes []Node
	Edges []*Edge
}

func (g *Graph) AddEdge(from, to NodeID, port string) {
	g.Edges = append(g.Edges, &Edge{
		From:    from,
		To:      to,
		ToPort:  port,
		Channel: make(chan []float64),
	})
}

func (g *Graph) AddGeneratorNode(gen generator.SampleGenerator, opts ...NodeOption) NodeID {
	var options nodeOptions
	for _, opt := range opts {
		opt(&options)
	}

	node := &GeneratorNode{
		id:        NodeID(len(g.Nodes)),
		Generator: gen,
		label:     options.label,
	}
	g.Nodes = append(g.Nodes, node)
	return node.id
}

func (g *Graph) AddSinkNode(opts ...NodeOption) (NodeID, <-chan []float64) {
	var options nodeOptions
	for _, opt := range opts {
		opt(&options)
	}

	node := &SinkNode{
		id:     NodeID(len(g.Nodes)),
		Output: make(chan []float64),
		label:  options.label,
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
	inputSamples := make(map[string][]float64)
	for _, e := range inEdges {
		select {
		case inputSamples[e.ToPort] = <-e.Channel:
		case <-ctx.Done():
			return
		}
	}

	cfg.InputSamples = inputSamples
	outputSamples := node.GenerateSamples(ctx, cfg, numSamples)

	for _, e := range g.OutgoingEdges(node.ID()) {
		select {
		case e.Channel <- outputSamples:
		case <-ctx.Done():
			return
		}
	}
}

// Dot returns a string representation of the graph in the DOT language.
func (g *Graph) Dot() string {
	var dot string
	dot += "digraph {\n"
	for _, node := range g.Nodes {
		// add node with its String() representation as label
		dot += fmt.Sprintf("\t%q [label=%q];\n", node.ID().String(), node.String())
	}
	for _, e := range g.Edges {
		// add an edge from e.From to e.To, using the node's ID as the
		// label.
		dot += "\t"
		dot += g.Nodes[e.From].ID().String()
		dot += " -> "
		dot += g.Nodes[e.To].ID().String()
		dot += "\n"
	}
	dot += "}\n"
	return dot
}
