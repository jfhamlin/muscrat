package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/jfhamlin/muscrat/pkg/bufferpool"
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

var (
	debugMode = os.Getenv("MUSCRAT_DEBUG") != "" && os.Getenv("MUSCRAT_DEBUG") != "0"
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
		epoch   atomic.Int64 // the _next_ epoch in which this node will be evaluated
		evaling atomic.Bool
		value   []float64 // value of the node in the last epoch

		incomingEdges []*Edge // edges whose destination is this node

		predecessors []NodeID // nodes that must be evaluated before this node

		// offset to apply to the epoch of each incoming edge when
		// checking if the dependency has been satisfied. this is
		// necessary to handle cycles in the graph.
		predecessorEpochOffsets []int64
	}

	// we evaluate the graph in epochs, where each epoch is a buffer
	// sent to all sinks.
	//
	// every epoch, we evaluate the graph in topological order using multiple
	// goroutines. each node has an atomic value containing the epoch number
	// of the last time it was evaluated. if the epoch number is less than
	// the current epoch, we evaluate the node and update the epoch number.
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

func (rs *runState) NodeInfoByIndex(idx int) *runNodeInfo {
	return &rs.nodeInfo[idx]
}

func (g *Graph) Run(ctx context.Context, cfg ugen.SampleConfig) {
	if g.BufferSize <= 0 {
		g.BufferSize = 1024
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
	if numWorkers > len(rs.nodeOrder) {
		numWorkers = len(rs.nodeOrder)
	}

	// var wg sync.WaitGroup
	// wg.Add(numWorkers)

	g.bootstrapCycles(ctx, rs)

	if debugMode {
		// print edges of nodes in the order
		for _, nodeID := range rs.nodeOrder {
			node := g.Node(nodeID)
			info := rs.NodeInfoByID(nodeID)
			fmt.Printf("node %d (%T) %s\n", nodeID, node, node)
			for i, pid := range info.predecessors {
				offset := info.predecessorEpochOffsets[i]
				fmt.Printf("  - %d <- %d (offset=%d)\n", nodeID, pid, offset)
			}
		}
	}

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
			g.runNode(ctx, cfg, rs, nodeID)
		}
	}

	q := NewQueue(numWorkers)
	items := make([]*QueueItem, len(rs.nodeOrder))
	for nodeIndex, nodeID := range rs.nodeOrder {
		info := rs.NodeInfoByIndex(nodeIndex)

		var numPreds int
		for _, predOffset := range info.predecessorEpochOffsets {
			if predOffset == 0 {
				numPreds++
			}
		}
		items[nodeIndex] = q.AddItem(makeRunFunc(nodeID), numPreds)
	}
	for nodeIndex, item := range items {
		info := rs.NodeInfoByIndex(nodeIndex)
		for i, predOffset := range info.predecessorEpochOffsets {
			if predOffset != 0 {
				continue
			}
			predIndex := rs.nodeIndexMap[info.predecessors[i]]
			predItem := items[predIndex]
			predItem.AddSuccessor(item)
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		q.Run(ctx)
	}

	// for i := 0; i < numWorkers; i++ {
	// 	id := i
	// 	go func() {
	// 		defer wg.Done()
	// 		g.runWorker(ctx, cfg, rs, id)
	// 	}()
	// }

	//wg.Wait()
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

	fullSampleSize := g.BufferSize * len(order)
	bufSlice := make([]float64, fullSampleSize, fullSampleSize)
	for i, id := range order {
		predecessorNodes := map[NodeID]struct{}{}
		incomingEdges := g.IncomingEdges(id)
		rs.nodeInfo[i].incomingEdges = incomingEdges
		for _, e := range g.IncomingEdges(id) {
			predecessorNodes[e.From] = struct{}{}
		}
		rs.nodeInfo[i].predecessors = make([]NodeID, 0, len(predecessorNodes))
		for pid := range predecessorNodes {
			rs.nodeInfo[i].predecessors = append(rs.nodeInfo[i].predecessors, pid)
		}
		rs.nodeInfo[i].predecessorEpochOffsets = make([]int64, len(predecessorNodes))
		rs.nodeInfo[i].value = bufSlice[i*g.BufferSize : (i+1)*g.BufferSize]
		rs.nodeIndexMap[id] = i
	}

	return rs
}

func (g *Graph) runWorker(ctx context.Context, cfg ugen.SampleConfig, rs *runState, workerID int) {
	var epoch int64 = 0 // the epoch being evaluated by this worker

	// the minimum epoch across all nodes
	// we can continue once this is > epoch
	var minEpoch int64 = 0

	inputSampleMap := make(map[string][]float64) // cleared after every node evaluation
Outer:
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if debugMode {
			fmt.Println(strings.Repeat("=", 80))
			fmt.Printf("- [%d] worker epoch %d\n", workerID, epoch)
			fmt.Printf("  - [%d] epoch:  %d\n", workerID, epoch)
			fmt.Printf("  - [%d] minEpoch: %d\n", workerID, minEpoch)
			for _, nodeID := range rs.nodeOrder {
				i := rs.nodeIndexMap[nodeID]
				info := rs.NodeInfoByIndex(i)
				fmt.Printf("  - [%d] node (%d %T) %d:\t%d\n", workerID, nodeID, g.Node(nodeID), i, info.epoch.Load())
			}
		}

		nextMinEpoch := int64(math.MaxInt64)

	NodeLoop:
		for nodeIndex, nodeID := range rs.nodeOrder {
			if minEpoch > epoch {
				epoch++
				continue Outer
			}
			node := g.Node(nodeID)
			info := rs.NodeInfoByIndex(nodeIndex)

			// if the node has already been evaluated for this epoch, skip it
			{
				nodeEpoch := info.epoch.Load()
				if nodeEpoch < nextMinEpoch {
					nextMinEpoch = nodeEpoch
				}
				if nodeEpoch > epoch {
					continue
				}
			}

			// try to lock the node for evaluation
			if !info.evaling.CompareAndSwap(false, true) {
				continue
			}

			//////////////////////////////////////////////////////////////////////
			// Node locked

			// check again if the node has already been evaluated for this
			// epoch after we've acquired the lock.
			if info.epoch.Load() > epoch {
				info.evaling.Store(false)
				continue
			}

			// first, check that all predecessors have been evaluated for
			// this epoch
			for i, pid := range info.predecessors {
				if rs.NodeInfoByID(pid).epoch.Load() <= epoch+info.predecessorEpochOffsets[i] {
					if debugMode {
						fmt.Printf("  - skipping eval for node %d, waiting\n", nodeID)
						fmt.Printf("    - for: %v offset: %v %v\n", pid, info.predecessorEpochOffsets[i], info.predecessorEpochOffsets)
					}
					info.evaling.Store(false)
					continue NodeLoop
				}
			}

			///////////////////////
			// evaluate node nodeID

			// then, collect the inputs
			for _, e := range info.incomingEdges {
				inputSampleMap[e.ToPort] = rs.NodeInfoByID(e.From).value
			}

			// process the node
			cfg.InputSamples = inputSampleMap
			switch n := node.(type) {
			case *GeneratorNode:
				clear(info.value)
				n.GenerateSamples(ctx, cfg, info.value)
				// update the epoch
				info.epoch.Store(epoch + 1)
			case *OutNode:
				for _, smps := range inputSampleMap { // TODO: disallow sinks with multiple inputs
					out := bufferpool.Get(len(smps))
					copy(*out, smps)
					// early epoch increment. we've copied the input samples, so
					// we can increment the epoch now. this allows the next
					// epoch to start while we wait on sending the samples to
					// the sink.
					info.epoch.Store(epoch + 1)
					select {
					case n.output <- *out:
					case <-ctx.Done():
						return
					}
					break
				}
			}

			// release the node
			info.evaling.Store(false)
			// Node unlocked
			//////////////////////////////////////////////////////////////////////

			clear(inputSampleMap)
		}

		minEpoch = nextMinEpoch
	}
}

func (g *Graph) runNode(ctx context.Context, cfg ugen.SampleConfig, rs *runState, nodeID NodeID) {
	inputSampleMap := make(map[string][]float64)

	info := rs.NodeInfoByID(nodeID)
	node := g.Node(nodeID)

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
		for _, smps := range inputSampleMap { // TODO: disallow sinks with multiple inputs
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

func (g *Graph) bootstrapCycles(ctx context.Context, rs *runState) {
	for _, sink := range g.Sinks() {
		visited := make(map[NodeID]struct{})
		g.prepCyclesDFS(rs, sink.ID(), visited)
	}
}

func (g *Graph) prepCyclesDFS(rs *runState, nodeID NodeID, visited map[NodeID]struct{}) {
	visited[nodeID] = struct{}{}
	defer delete(visited, nodeID)

	info := rs.NodeInfoByID(nodeID)
	for i, from := range info.predecessors {
		if _, ok := visited[from]; ok {
			info.predecessorEpochOffsets[i] = -1
			continue
		}
		g.prepCyclesDFS(rs, from, visited)
	}
}
