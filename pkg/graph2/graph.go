package graph2

import (
	"github.com/glojurelang/glojure/pkg/lang"
)

type (
	NodeID string

	Graph struct {
		Nodes []*Node
		Edges []*Edge
	}

	Node struct {
		ID   NodeID
		Type string
		Args any
		Key  string
		Sink bool
	}

	Edge struct {
		From NodeID
		To   NodeID
		Port string
	}

	// GraphAlignment is a struct that represents the alignment of two
	// graphs.
	GraphAlignment struct {
		NodeIdentities map[NodeID]NodeID
	}
)

var (
	typeKW  = lang.NewKeyword("type")
	outKW   = lang.NewKeyword("out")
	argsKW  = lang.NewKeyword("args")
	ctorKW  = lang.NewKeyword("ctor")
	idKW    = lang.NewKeyword("id")
	sinkKW  = lang.NewKeyword("sink")
	constKW = lang.NewKeyword("const")
	nodesKW = lang.NewKeyword("nodes")
	edgesKW = lang.NewKeyword("edges")
	fromKW  = lang.NewKeyword("from")
	toKW    = lang.NewKeyword("to")
	portKW  = lang.NewKeyword("port")
	keyKW   = lang.NewKeyword("key")
)

func SExprToGraph(sexpr any) *Graph {
	g := &Graph{}
	nodes := lang.Get(sexpr, nodesKW)
	edges := lang.Get(sexpr, edgesKW)

	for s := lang.Seq(nodes); s != nil; s = lang.Next(s) {
		node := lang.First(s)
		id, _ := lang.Get(node, idKW).(string)
		typ, _ := lang.Get(node, typeKW).(lang.Keyword)
		args := lang.Get(node, argsKW)
		key, _ := lang.Get(node, keyKW).(string)
		sink, _ := lang.Get(node, sinkKW).(bool)
		g.Nodes = append(g.Nodes, &Node{
			ID:   NodeID(id),
			Type: typ.Name(),
			Args: args,
			Key:  key,
			Sink: sink,
		})
	}

	for s := lang.Seq(edges); s != nil; s = lang.Next(s) {
	}

	return g
}

func (g *Graph) Sinks() []*Node {
	var sinks []*Node
	for _, n := range g.Nodes {
		if n.Sink {
			sinks = append(sinks, n)
		}
	}
	return sinks
}

func (g *Graph) Node(id NodeID) *Node {
	for _, n := range g.Nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}
