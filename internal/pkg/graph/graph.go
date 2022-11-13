package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strconv"
	"sync"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
)

const (
	bufferSize = 1024
)

type NodeID int

func (id NodeID) String() string {
	return strconv.Itoa(int(id))
}

type SinkChan <-chan []float64

type Node interface {
	ID() NodeID
	Run(ctx context.Context, g *Graph, cfg generator.SampleConfig, numSamples int)

	String() string

	json.Marshaler
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

func (n *GeneratorNode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"id":%d,"type":"generator","label":"%s"}`, n.id, n.String())), nil
}

type SinkNode struct {
	id     NodeID
	output chan []float64
	label  string
}

func (n *SinkNode) ID() NodeID {
	return n.id
}

func (n *SinkNode) Chan() SinkChan {
	return n.output
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
		case n.output <- inputSamples[0]:
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

func (n *SinkNode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"id":%d,"type":"sink","label":"%s"}`, n.id, n.String())), nil
}

type Edge struct {
	From    NodeID
	To      NodeID
	ToPort  string
	Channel chan []float64
}

func (e *Edge) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"from":%d,"to":%d,"toPort":"%s"}`, e.From, e.To, e.ToPort)), nil
}

// Graph is a graph of SampleGenerators.
type Graph struct {
	Nodes []Node  `json:"nodes"`
	Edges []*Edge `json:"edges"`
}

func (g *Graph) AddEdge(from, to NodeID, port string) {
	g.Edges = append(g.Edges, &Edge{
		From:    from,
		To:      to,
		ToPort:  port,
		Channel: make(chan []float64, 1),
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

func (g *Graph) AddSinkNode(opts ...NodeOption) *SinkNode {
	var options nodeOptions
	for _, opt := range opts {
		opt(&options)
	}

	node := &SinkNode{
		id:     NodeID(len(g.Nodes)),
		output: make(chan []float64),
		label:  options.label,
	}
	g.Nodes = append(g.Nodes, node)
	return node
}

func (g *Graph) Sinks() []*SinkNode {
	var sinks []*SinkNode
	for _, node := range g.Nodes {
		if sink, ok := node.(*SinkNode); ok {
			sinks = append(sinks, sink)
		}
	}
	return sinks
}

func (g *Graph) Node(id NodeID) Node {
	if int(id) >= len(g.Nodes) {
		return nil
	}
	return g.Nodes[id]
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
	bootstrapCycles(ctx, g, cfg)

	var wg sync.WaitGroup
	for _, node := range g.Nodes {
		wg.Add(1)
		go func(n Node) {
			n.Run(ctx, g, cfg, bufferSize)
			wg.Done()
		}(node)
	}
	wg.Wait()
}

func bootstrapCycles(ctx context.Context, g *Graph, cfg generator.SampleConfig) {
	// initialize any channels required to bootstrap cycles, preventing
	// deadlock.

	queue := make([]NodeID, 0, len(g.Nodes))
	blocked := map[NodeID]struct{}{}
	for _, node := range g.Nodes {
		if _, ok := node.(*SinkNode); ok {
			continue
		}
		if len(g.IncomingEdges(node.ID())) == 0 {
			queue = append(queue, node.ID())
		} else {
			blocked[node.ID()] = struct{}{}
		}
	}

	satisfiedNodes := make(map[NodeID]struct{})
	for len(queue) > 0 || len(blocked) > 0 {
		if len(queue) == 0 {
			// all nodes are blocked. pick an unsatisfied dependency of the
			// node with the most satisfied dependencies, and treat it as
			// "satisfied," generating a buffer of zero samples for it.
			maxSatisfied := -1
			var choice NodeID = -1

			for id := range blocked {
				satisfiedDeps := 0
				var unsatisfiedDep NodeID
				for _, e := range g.IncomingEdges(id) {
					if _, ok := satisfiedNodes[e.From]; ok {
						satisfiedDeps++
					} else {
						unsatisfiedDep = e.From
					}
				}
				if satisfiedDeps > maxSatisfied {
					maxSatisfied = satisfiedDeps
					choice = unsatisfiedDep
				}
			}

			queue = append(queue, choice)
			delete(blocked, choice)
			zeros := make([]float64, bufferSize)
			for _, e := range g.OutgoingEdges(choice) {
				select {
				case <-ctx.Done():
					return
				case e.Channel <- zeros:
				}
			}
		}

		var nodeID NodeID
		nodeID, queue = queue[0], queue[1:]

		satisfiedNodes[nodeID] = struct{}{}
		// check for any unblocked nodes
		for _, e := range g.OutgoingEdges(nodeID) {
			if _, ok := satisfiedNodes[e.To]; ok {
				continue
			}
			satisfied := true
			for _, inEdge := range g.IncomingEdges(e.To) {
				if _, ok := satisfiedNodes[inEdge.From]; !ok {
					satisfied = false
					break
				}
			}
			if !satisfied {
				continue
			}
			delete(blocked, e.To)
			queue = append(queue, e.To)
		}
	}
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
		// add an edge from e.From to e.To, using e.ToPort as label
		dot += fmt.Sprintf("\t%q -> %q [label=%q];\n", e.From.String(), e.To.String(), e.ToPort)
	}
	dot += "}\n"
	return dot
}
