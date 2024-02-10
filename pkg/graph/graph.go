package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"

	"github.com/jfhamlin/muscrat/pkg/bufferpool"
	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

// TODO: prune nodes that aren't connected to any sinks. NB that some
// nodes today aren't connected to sinks but have side effects; we
// need a way to identify those nodes.

type (
	NodeID int

	SinkChan <-chan []float64

	nodeOptions struct {
		label  string
		isSink bool
	}

	// NodeOptions is a functional option for configuring a node.
	NodeOption func(*nodeOptions)

	Node interface {
		ID() NodeID

		IsSink() bool

		String() string

		json.Marshaler
	}

	GeneratorNode struct {
		id        NodeID
		Generator ugen.UGen
		label     string
		str       string

		// last output
		value []float64

		isSink bool
	}

	OutNode struct {
		id     NodeID
		output chan []float64
		label  string
	}

	Edge struct {
		From   NodeID
		To     NodeID
		ToPort string
	}

	// Graph is a graph of SampleGenerators.
	Graph struct {
		Nodes []Node  `json:"nodes"`
		Edges []*Edge `json:"edges"`

		BufferSize int `json:"bufferSize"`

		outputs []*OutNode
	}
)

func (id NodeID) String() string {
	return strconv.Itoa(int(id))
}

func WithLabel(label string) NodeOption {
	return func(o *nodeOptions) {
		o.label = label
	}
}

func WithSink() NodeOption {
	return func(o *nodeOptions) {
		o.isSink = true
	}
}

func (n *GeneratorNode) ID() NodeID {
	return n.id
}

func (n *GeneratorNode) IsSink() bool {
	return n.isSink
}

func (n *GeneratorNode) GenerateSamples(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	defer func() {
		if r := recover(); r != nil {
			// TODO: make the failure of this node visible to the user.
			//fmt.Printf("node %s failed: %v\n", n, r)
			// print stack trace
			fmt.Printf("node %s failed: %v\n", n, r)
			fmt.Printf("stack trace: %s\n", debug.Stack())
			clear(out)
		}
	}()
	n.Generator.Gen(ctx, cfg, out)
}

func (n *GeneratorNode) String() string {
	return n.str
}

func (n *GeneratorNode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"id":%d,"type":"generator","label":"%s"}`, n.id, n.String())), nil
}

func (n *OutNode) ID() NodeID {
	return n.id
}

func (n *OutNode) IsSink() bool {
	return true
}

func (n *OutNode) Chan() SinkChan {
	return n.output
}

func (n *OutNode) String() string {
	if n.label != "" {
		return n.label
	}
	return n.ID().String()
}

func (n *OutNode) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"id":%d,"type":"sink","label":"%s"}`, n.id, n.String())), nil
}

func (e *Edge) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{"from":%d,"to":%d,"toPort":"%s"}`, e.From, e.To, e.ToPort)), nil
}

func (g *Graph) AddEdge(from, to NodeID, port string) {
	// panic on edges whose source or destination is not in the graph,
	// or whose source is a sink
	if int(from) >= len(g.Nodes) {
		panic(fmt.Sprintf("edge source %d is not in the graph", from))
	}
	if int(to) >= len(g.Nodes) {
		panic(fmt.Sprintf("edge destination %d is not in the graph", to))
	}
	if from == to {
		panic(fmt.Sprintf("edge source %d and destination %d are the same", from, to))
	}
	if _, ok := g.Nodes[from].(*OutNode); ok {
		panic(fmt.Sprintf("cannot add edge whose source %d is a sink node", from))
	}

	g.Edges = append(g.Edges, &Edge{
		From:   from,
		To:     to,
		ToPort: port,
	})
}

func (g *Graph) AddGeneratorNode(gen ugen.UGen, opts ...NodeOption) Node {
	var options nodeOptions
	for _, opt := range opts {
		opt(&options)
	}

	node := &GeneratorNode{
		id:        NodeID(len(g.Nodes)),
		Generator: gen,
		label:     options.label,
		isSink:    options.isSink,
		value:     make([]float64, g.BufferSize),
	}
	{
		str := "[" + node.ID().String() + "]"
		if node.label != "" {
			str += " " + node.label
		}
		node.str = str
	}

	g.Nodes = append(g.Nodes, node)
	return node
}

func (g *Graph) AddOutNode(opts ...NodeOption) *OutNode {
	var options nodeOptions
	for _, opt := range opts {
		opt(&options)
	}

	node := &OutNode{
		id:     NodeID(len(g.Nodes)),
		output: make(chan []float64),
		label:  options.label,
	}
	g.Nodes = append(g.Nodes, node)
	g.outputs = append(g.outputs, node)
	return node
}

func (g *Graph) Sinks() []Node {
	var sinks []Node
	for _, node := range g.Nodes {
		if node.IsSink() {
			sinks = append(sinks, node)
		}
	}
	return sinks
}

func (g *Graph) Outputs() []*OutNode {
	return g.outputs
}

func (g *Graph) OutputChans() []SinkChan {
	var chs []SinkChan
	for _, out := range g.outputs {
		chs = append(chs, out.output)
	}
	return chs
}

func (g *Graph) Node(id NodeID) Node {
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

////////////////////////////////////////////////////////////////////////////////
// Graph Running

type (
	runNodeInfo struct {
		value []float64

		// edges whose destination is this node
		incomingEdges []*Edge

		// nodes that must be evaluated before this node. not all
		// predecessors are made dependencies to eliminate dependency
		// cycles.
		dependencies map[NodeID]struct{}

		node Node
	}

	runState struct {
		nodeInfo []runNodeInfo

		// nodeOrder is a topological-ish ordering of the nodes in the
		// graph, excluding nodes that are not ancestors of any
		// sink. Because the graphs can contain cycles, this ordering is
		// not guaranteed to be a topological ordering.
		nodeOrder    []NodeID
		nodeIndexMap []int
	}
)

func (rs *runState) NodeInfoByID(id NodeID) *runNodeInfo {
	index := rs.nodeIndexMap[id]
	if index < 0 {
		return nil
	}
	return &rs.nodeInfo[index]
}

func (g *Graph) Run(ctx context.Context, cfg ugen.SampleConfig) {
	if g.BufferSize <= 0 {
		g.BufferSize = conf.BufferSize
	}

	// using 1/2 the number of CPUs gives good performance when
	// benchmarking. using all CPUs gives worse performance.
	numWorkers := runtime.NumCPU() / 2
	if workerEnvVar := os.Getenv("MUSCRAT_WORKERS"); workerEnvVar != "" {
		if n, err := strconv.Atoi(workerEnvVar); err == nil {
			numWorkers = n
		}
	}
	if numWorkers < 1 {
		numWorkers = 1
	}

	rs := g.newRunState()

	// start any generator nodes whose generator is a ugen.Starter
	for _, node := range g.Nodes {
		if gen, ok := node.(*GeneratorNode); ok {
			if starter, ok := gen.Generator.(ugen.Starter); ok {
				starter.Start(ctx)
			}
		}
	}
	// defer stopping any generator nodes whose generator is a ugen.Stopper
	defer func() {
		for _, node := range g.Nodes {
			if gen, ok := node.(*GeneratorNode); ok {
				if stopper, ok := gen.Generator.(ugen.Stopper); ok {
					stopper.Stop(ctx)
				}
			}
		}
	}()

	makeRunFunc := func(nodeID NodeID) Job {
		return func(ctx context.Context) {
			runNode(ctx, cfg, rs, nodeID)
		}
	}

	q := NewQueue(numWorkers)
	items := map[NodeID]*QueueItem{}
	for _, nodeID := range rs.nodeOrder {
		items[nodeID] = q.AddItem(makeRunFunc(nodeID))
	}
	for nodeID, item := range items {
		info := rs.NodeInfoByID(nodeID)
		for depID := range info.dependencies {
			depItem := items[depID]
			depItem.AddSuccessor(item)
		}
	}

	q.Start(ctx)
	defer q.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := q.RunJobs(ctx); err != nil {
			break
		}
	}
}

func (g *Graph) newRunState() *runState {
	// build a topological-ish ordering of the nodes in the graph
	// (excluding nodes that are not ancestors of any sink). Because
	// the graphs can contain cycles, this ordering is not guaranteed
	// to be a topological ordering.
	//
	// we do this by starting with the sinks and working backwards
	// through the graph.
	var (
		visited = make([]bool, len(g.Nodes))
		order   []NodeID
	)

	var q []NodeID
	for _, sink := range g.Sinks() {
		q = append(q, sink.ID())
	}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		if visited[n] {
			continue
		}
		visited[n] = true
		order = append(order, n)
		for _, e := range g.IncomingEdges(n) {
			q = append(q, e.From)
		}
	}
	// reverse the order
	for i := 0; i < len(order)/2; i++ {
		j := len(order) - i - 1
		order[i], order[j] = order[j], order[i]
	}

	rs := &runState{
		nodeInfo:     make([]runNodeInfo, len(order)),
		nodeOrder:    order,
		nodeIndexMap: make([]int, len(g.Nodes)),
	}
	for i := range rs.nodeIndexMap {
		// initialize to -1 so that we can detect nodes that are not
		// ancestors of any sink.
		rs.nodeIndexMap[i] = -1
	}

	for i, id := range order {
		info := &rs.nodeInfo[i]

		predecessorNodes := map[NodeID]struct{}{}
		incomingEdges := g.IncomingEdges(id)
		info.incomingEdges = incomingEdges
		for _, e := range g.IncomingEdges(id) {
			predecessorNodes[e.From] = struct{}{}
		}
		info.dependencies = predecessorNodes
		node := g.Node(id)
		info.node = node
		if gen, ok := node.(*GeneratorNode); ok {
			info.value = gen.value
		}
		rs.nodeIndexMap[id] = i
	}

	bootstrapCycles(rs)

	return rs
}

func runNode(ctx context.Context, cfg ugen.SampleConfig, rs *runState, nodeID NodeID) {
	inputSampleMap := make(map[string][]float64)

	info := rs.NodeInfoByID(nodeID)
	node := info.node

	// collect the inputs
	for _, e := range info.incomingEdges {
		inputSampleMap[e.ToPort] = rs.NodeInfoByID(e.From).value
	}

	// process the node
	cfg.InputSamples = inputSampleMap
	switch n := node.(type) {
	case *GeneratorNode:
		clear(info.value)
		n.GenerateSamples(ctx, cfg, info.value)
	case *OutNode:
		for _, smps := range inputSampleMap { // TODO: disallow sinks with multiple inputs; or sum inputs
			out := bufferpool.Get(len(smps))
			copy(*out, smps)
			select {
			case n.output <- *out:
			case <-ctx.Done():
				return
			}
			break
		}
	}
}

func bootstrapCycles(rs *runState) {
	for _, info := range rs.nodeInfo {
		if info.node.IsSink() {
			visited := make(map[NodeID]struct{})
			prepCyclesDFS(rs, info.node.ID(), visited)
		}
	}
}

func prepCyclesDFS(rs *runState, nodeID NodeID, visited map[NodeID]struct{}) {
	visited[nodeID] = struct{}{}
	defer delete(visited, nodeID)

	info := rs.NodeInfoByID(nodeID)
	for from := range info.dependencies {
		if _, ok := visited[from]; ok {
			// cycle detected
			// remove the dependency
			delete(info.dependencies, from)
			continue
		}
		prepCyclesDFS(rs, from, visited)
	}
}
