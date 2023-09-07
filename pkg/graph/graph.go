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
	"github.com/jfhamlin/muscrat/pkg/prof"
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
		Run(ctx context.Context, g *Graph, cfg ugen.SampleConfig, numSamples int)

		String() string

		json.Marshaler
	}

	GeneratorNode struct {
		id        NodeID
		Generator ugen.UGen
		label     string
		str       string
	}

	SinkNode struct {
		id     NodeID
		output chan []float64
		label  string

		sinkID int
	}

	edgeVal struct {
		epoch   int64
		samples *[]float64
		waiters atomic.Int32
	}

	Edge struct {
		From   NodeID
		To     NodeID
		ToPort string

		channel chan *edgeVal
	}

	// Graph is a graph of SampleGenerators.
	Graph struct {
		Nodes []Node  `json:"nodes"`
		Edges []*Edge `json:"edges"`

		BufferSize int `json:"bufferSize"`

		sinks []*SinkNode
	}
)

var (
	edgeValPool = sync.Pool{
		New: func() any {
			return &edgeVal{}
		},
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

func (n *GeneratorNode) Run(ctx context.Context, g *Graph, cfg ugen.SampleConfig, numSamples int) {
	incomingEdges := g.IncomingEdges(n.id)
	outgoingEdges := g.OutgoingEdges(n.id)
	if len(incomingEdges) == 0 && len(outgoingEdges) == 0 {
		// this node is not connected to anything, so there's nothing to do
		fmt.Printf("node %s is not connected to anything\n", n)
		return
	}

	if starter, ok := n.Generator.(ugen.Starter); ok {
		if err := starter.Start(ctx); err != nil {
			panic(err)
		}
	}
	defer func() {
		if stopper, ok := n.Generator.(ugen.Stopper); ok {
			if err := stopper.Stop(ctx); err != nil {
				panic(err)
			}
		}
	}()

	inputSamples := make(map[string][]float64)
	inputEdgeVals := make([]*edgeVal, 0, len(incomingEdges))
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		for _, e := range incomingEdges {
			select {
			case val := <-e.channel:
				inputSamples[e.ToPort] = *val.samples
				inputEdgeVals = append(inputEdgeVals, val)
			case <-ctx.Done():
				return
			}
		}

		span := prof.StartSpan(ctx, n.String())
		cfg.InputSamples = inputSamples
		ev := newEdgeVal(numSamples, len(outgoingEdges))
		n.GenerateSamples(ctx, cfg, *ev.samples)
		span.Finish()
		// signal to the input edge vals that we're done with them
		for _, ev := range inputEdgeVals {
			ev.done()
		}
		inputEdgeVals = inputEdgeVals[:0]

		for _, e := range outgoingEdges {
			select {
			case e.channel <- ev:
			case <-ctx.Done():
				return
			}
		}
	}
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

func (n *SinkNode) ID() NodeID {
	return n.id
}

func (n *SinkNode) Chan() SinkChan {
	return n.output
}

func (n *SinkNode) Run(ctx context.Context, g *Graph, cfg ugen.SampleConfig, numSamples int) {
	inEdges := g.IncomingEdges(n.id)
	inputEdgeVals := make([]*edgeVal, 0, len(inEdges))
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		//start := time.Now()

		inputSamples := make([][]float64, len(inEdges))
		for i, e := range inEdges {
			select {
			case <-ctx.Done():
				return
			case val := <-e.channel:
				out := bufferpool.Get(len(*val.samples))
				copy(*out, *val.samples)
				inputSamples[i] = *out
				inputEdgeVals = append(inputEdgeVals, val)
			}
		}
		if len(inputSamples) == 0 {
			continue
		}
		// if dur := time.Since(start); dur > time.Duration(numSamples)*time.Second/time.Duration(cfg.SampleRateHz) {
		// 	fmt.Printf("[SLOW] got samples in %s\n", time.Since(start))
		// }
		select {
		case n.output <- inputSamples[0]:
		case <-ctx.Done():
			return
		}
		for _, ev := range inputEdgeVals {
			ev.done()
		}
		inputEdgeVals = inputEdgeVals[:0]
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
	if _, ok := g.Nodes[from].(*SinkNode); ok {
		panic(fmt.Sprintf("cannot add edge whose source %d is a sink node", from))
	}

	g.Edges = append(g.Edges, &Edge{
		From:    from,
		To:      to,
		ToPort:  port,
		channel: make(chan *edgeVal, 1),
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

func (g *Graph) AddSinkNode(opts ...NodeOption) *SinkNode {
	var options nodeOptions
	for _, opt := range opts {
		opt(&options)
	}

	node := &SinkNode{
		id:     NodeID(len(g.Nodes)),
		output: make(chan []float64),
		label:  options.label,
		sinkID: len(g.sinks),
	}
	g.Nodes = append(g.Nodes, node)
	g.sinks = append(g.sinks, node)
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

func (g *Graph) SinkChans() []SinkChan {
	var chs []SinkChan
	for _, node := range g.Nodes {
		if sink, ok := node.(*SinkNode); ok {
			chs = append(chs, sink.output)
		}
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

func (g *Graph) Run(ctx context.Context, cfg ugen.SampleConfig) {
	if g.BufferSize <= 0 {
		g.BufferSize = 1024
	}

	g.bootstrapCycles(ctx)

	var wg sync.WaitGroup
	for _, node := range g.Nodes {
		wg.Add(1)
		go func(n Node) {
			n.Run(ctx, g, cfg, g.BufferSize)
			wg.Done()
		}(node)
	}
	wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
// exploratory

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
	return &rs.nodeInfo[rs.nodeIndexMap[id]]
}

func (rs *runState) NodeInfoByIndex(idx int) *runNodeInfo {
	return &rs.nodeInfo[idx]
}

func (g *Graph) RunWorkers(ctx context.Context, cfg ugen.SampleConfig) {
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
	fmt.Printf("XXX running with %d workers (%d nodes)\n", numWorkers, len(rs.nodeOrder))

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

	numSinks := len(g.Sinks())
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
				if sink, ok := node.(*SinkNode); ok {
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
				if sink, ok := node.(*SinkNode); ok {
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
			case *SinkNode:
				for _, ev := range info.incomingEdges { // TODO: disallow sinks with multiple inputs
					smps := rs.NodeInfoByID(ev.From).value
					out := bufferpool.Get(len(smps))
					copy(*out, smps)
					select {
					case n.output <- *out:
					case <-ctx.Done():
						return
					}
					break
				}
				sinkDoneMask |= 1 << n.sinkID
			}

			// update the epoch
			info.epoch.Store(epoch + 1)

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
			rs.NodeInfoByID(choice).epoch.Store(1)
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

// end exploratory
////////////////////////////////////////////////////////////////////////////////

func (g *Graph) bootstrapCycles(ctx context.Context) {
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
			zeros := newEdgeVal(g.BufferSize, len(g.OutgoingEdges(choice)))
			for _, e := range g.OutgoingEdges(choice) {
				select {
				case <-ctx.Done():
					return
				case e.channel <- zeros:
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

func newEdgeVal(numSamples, waiters int) *edgeVal {
	ev := edgeValPool.Get().(*edgeVal)
	ev.samples = bufferpool.Get(numSamples)
	ev.waiters.Store(int32(waiters))
	return ev
}

func (ev *edgeVal) done() int32 {
	dec := ev.waiters.Add(-1)
	if dec <= 0 {
		bufferpool.Put(ev.samples)
		edgeValPool.Put(ev)
	}
	return dec
}
