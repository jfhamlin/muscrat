package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
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
		label string
	}

	// NodeOptions is a functional option for configuring a node.
	NodeOption func(*nodeOptions)

	Node interface {
		ID() NodeID

		Sink() bool

		String() string

		json.Marshaler
	}

	GeneratorNode struct {
		id        NodeID
		Generator ugen.UGen
		label     string
		str       string

		sinkID int
	}

	OutNode struct {
		id     NodeID
		output chan []float64
		label  string

		sinkID int
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

func (n *GeneratorNode) ID() NodeID {
	return n.id
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
		sinkID: len(g.outputs),
	}
	g.Nodes = append(g.Nodes, node)
	g.outputs = append(g.outputs, node)
	return node
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

////////////////////////////////////////////////////////////////////////////////
// Graph Running

type (
	runNodeInfo struct {
		epoch         atomic.Int64
		evaling       atomic.Bool
		value         []float64 // value of the node in the last epoch
		incomingEdges []*Edge
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
	if numWorkers < 1 {
		numWorkers = 1
	}

	rs := g.newRunState()
	if numWorkers > len(rs.nodeOrder) {
		numWorkers = len(rs.nodeOrder)
	}

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	g.bootstrapCyclesWorkers(ctx, rs)

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

	for i := 0; i < numWorkers; i++ {
		id := i
		go func() {
			defer wg.Done()
			g.runWorker(ctx, cfg, rs, id)
		}()
	}
	wg.Wait()
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
	for _, sink := range g.Outputs() {
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
		rs.nodeInfo[i].incomingEdges = g.IncomingEdges(id)
		rs.nodeInfo[i].value = bufSlice[i*g.BufferSize : (i+1)*g.BufferSize]
		rs.nodeIndexMap[id] = i
	}

	return rs
}

func (g *Graph) runWorker(ctx context.Context, cfg ugen.SampleConfig, rs *runState, workerID int) {
	// invariants:
	// 1. rs.nodeEpochs[i] is the next unevaluated epoch of node i
	// 2. rs.nodeEvaling[i] is true if node i is currently being evaluated
	// 3. rs.nodeValues[i] is the value of node i for the previous epoch of node i

	numSinks := len(g.Outputs())
	sinksDoneBits := (int64(1) << numSinks) - 1

	var epoch int64 = 0    // the epoch being evaluated by this worker
	var sinkDoneMask int64 // assumes numSinks <= 64
	if numSinks > 64 {
		panic("too many sinks")
	}

	inputSampleMap := make(map[string][]float64) // cleared after every node evaluation
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// fmt.Printf("- [%d] worker epoch %d\n", workerID, epoch)
		// fmt.Printf("  - [%d] sinkDoneMask:  %064b\n", workerID, sinkDoneMask)
		// fmt.Printf("  - [%d] sinksDoneBits: %064b\n", workerID, sinksDoneBits)
		// for i := range rs.nodeEpochs {
		// 	fmt.Printf("  - [%d] node %d epoch: %d\n", workerID, i, rs.nodeEpochs[i].Load())
		// }

	NodeLoop:
		for nodeIndex, nodeID := range rs.nodeOrder {
			if sinkDoneMask == sinksDoneBits {
				epoch++
				sinkDoneMask = 0
			}
			node := g.Node(nodeID)
			info := rs.NodeInfoByIndex(nodeIndex)

			// if the node has already been evaluated for this epoch, skip it
			if info.epoch.Load() > epoch {
				if sink, ok := node.(*OutNode); ok {
					sinkDoneMask |= 1 << sink.sinkID
				}
				continue
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
				if sink, ok := node.(*OutNode); ok {
					sinkDoneMask |= 1 << sink.sinkID
				}
				info.evaling.Store(false)
				continue
			}

			// first, check that all nodes with incoming edges have been
			// evaluated for this epoch
			for _, e := range info.incomingEdges {
				if rs.NodeInfoByID(e.From).epoch.Load() <= epoch {
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
				sinkDoneMask |= 1 << n.sinkID
			}

			// release the node
			info.evaling.Store(false)
			// Node unlocked
			//////////////////////////////////////////////////////////////////////

			clear(inputSampleMap)
		}
	}
}

func (g *Graph) bootstrapCyclesWorkers(ctx context.Context, rs *runState) {
	// initialize any channels required to bootstrap cycles, preventing
	// deadlock.

	queue := make([]NodeID, 0, len(g.Nodes))
	blocked := map[NodeID]struct{}{}
	for _, node := range g.Nodes {
		if _, ok := node.(*OutNode); ok {
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
			if ni := rs.NodeInfoByID(choice); ni != nil {
				rs.NodeInfoByID(choice).epoch.Store(1)
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

//
////////////////////////////////////////////////////////////////////////////////

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
