package graph2

import (
	"github.com/glojurelang/glojure/pkg/lang"
	"github.com/jfhamlin/muscrat/pkg/ugen"
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
		Ctor any
		Args any
		Key  string
		Sink bool
	}

	Edge struct {
		From NodeID
		To   NodeID
		Port string
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
		ctor := lang.Get(node, ctorKW)
		args := lang.Get(node, argsKW)
		key, _ := lang.Get(node, keyKW).(string)
		sink, _ := lang.Get(node, sinkKW).(bool)
		g.Nodes = append(g.Nodes, &Node{
			ID:   NodeID(id),
			Type: typ.Name(),
			Ctor: ctor,
			Args: args,
			Key:  key,
			Sink: sink,
		})
	}

	for s := lang.Seq(edges); s != nil; s = lang.Next(s) {
		edge := lang.First(s)
		fromVal := lang.Get(edge, fromKW)
		from, ok := fromVal.(string)
		if !ok {
			panic("edge 'from' must be a string")
		}
		toVal := lang.Get(edge, toKW)
		to, ok := toVal.(string)
		if !ok {
			panic("edge 'to' must be a string")
		}
		portVal := lang.Get(edge, portKW)
		port, ok := portVal.(string)
		if !ok {
			panic("edge 'port' must be a string")
		}

		g.Edges = append(g.Edges, &Edge{
			From: NodeID(from),
			To:   NodeID(to),
			Port: port,
		})
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

func (n *Node) Construct() ugen.UGen {
	if n.Ctor == nil {
		return nil
	}

	res := lang.Apply(n.Ctor, seqToSlice(n.Args))
	u, ok := res.(ugen.UGen)
	if !ok {
		panic("node ctor must return a UGen")
	}
	return u
}
