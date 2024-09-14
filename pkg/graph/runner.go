package graph

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"

	"github.com/glojurelang/glojure/pkg/lang"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

type (
	Runner struct {
		ctx context.Context

		sampleConfig ugen.SampleConfig

		g  *Graph
		rs *runState

		nextID runNodeID

		epochChan chan runEpoch

		nextOut [][]float64

		out chan [][]float64

		mtx sync.Mutex
	}

	runEpoch struct {
		q      *Queue
		prevRS *runState
	}

	runNodeID int64

	runNode struct {
		id runNodeID

		gen ugen.UGen

		inputSampleMap map[string][]float64

		// edges whose destination is this node
		incomingEdges []*Edge

		// nodes that must be evaluated before this node. not all
		// predecessors are made dependencies to eliminate dependency
		// cycles.
		dependencies map[runNodeID]struct{}

		node *Node

		value []float64

		retained bool
	}

	runState struct {
		nodes []runNode

		// nodeOrder is a topological-ish ordering of the nodes in the
		// graph, excluding nodes that are not ancestors of any
		// sink. Because the graphs can contain cycles, this ordering is
		// not guaranteed to be a topological ordering.
		nodeOrder    []runNodeID
		nodeIndexMap map[runNodeID]int
	}
)

var (
	numWorkers = func() int {
		nw := runtime.NumCPU() / 2
		if workerEnvVar := os.Getenv("MUSCRAT_WORKERS"); workerEnvVar != "" {
			if n, err := strconv.Atoi(workerEnvVar); err == nil {
				nw = n
			}
		}
		if nw < 1 {
			nw = 1
		}
		return nw
	}()
)

func NewRunner(ctx context.Context, cfg ugen.SampleConfig, out chan [][]float64) *Runner {
	nextOut := make([][]float64, 2)
	for i := range nextOut {
		nextOut[i] = make([]float64, conf.BufferSize)
	}
	return &Runner{
		ctx:          ctx,
		sampleConfig: cfg,
		epochChan:    make(chan runEpoch),
		nextOut:      nextOut,
		out:          out,
	}
}

func (r *Runner) getNextID() runNodeID {
	r.nextID++
	return r.nextID
}

func (r *Runner) SetGraph(g *Graph) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	rs := r.newRunState(g)

	makeRunFunc := func(nid runNodeID) Job {
		node := rs.NodeByID(nid)
		return func(ctx context.Context) {
			node.run(ctx, r.sampleConfig)
		}
	}

	q := NewQueue(numWorkers)
	items := map[runNodeID]*QueueItem{}
	for _, nid := range rs.nodeOrder {
		items[nid] = q.AddItem(makeRunFunc(nid))
	}
	for nid, item := range items {
		node := rs.NodeByID(nid)
		for depID := range node.dependencies {
			depItem := items[depID]
			depItem.AddSuccessor(item)
		}
	}

	q.Start(r.ctx)

	r.g = g
	prevRS := r.rs
	r.rs = rs
	r.epochChan <- runEpoch{
		q:      q,
		prevRS: prevRS,
	}
}

func (r *Runner) Run(ctx context.Context) {
	q := NewQueue(1)
	q.Start(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case nxt := <-r.epochChan:
			go q.Stop()
			prevRS := nxt.prevRS
			if prevRS != nil {
				go func() {
					// stop nodes that are no longer in the graph
					for i := range prevRS.nodes {
						n := &prevRS.nodes[i]
						if n.retained {
							continue
						}
						if s, ok := n.gen.(ugen.Stopper); ok {
							s.Stop(ctx)
						}
					}
				}()
			}
			q = nxt.q
		default:
		}

		if err := q.RunJobs(ctx); err != nil {
			break
		}

		// copy the output buffer to the output channel
		outputBuffer := make([][]float64, len(r.nextOut))
		for i, out := range r.nextOut {
			outputBuffer[i] = make([]float64, len(out))
			copy(outputBuffer[i], out)
		}

		r.out <- outputBuffer
	}
}

func (r *Runner) newRunState(g *Graph) *runState {
	// build a topological-ish ordering of the nodes in the graph
	// (excluding nodes that are not ancestors of any sink). Because
	// the graphs can contain cycles, this ordering is not guaranteed
	// to be a topological ordering.
	//
	// we do this by starting with the sinks and working backwards
	// through the graph.
	var (
		visited       = make(map[runNodeID]bool)
		order         []runNodeID
		idMap         = map[NodeID]runNodeID{}
		nodeMap       = map[runNodeID]*Node{}
		incomingEdges = map[runNodeID][]*Edge{}
	)
	getID := func(id NodeID) runNodeID {
		if n, ok := idMap[id]; ok {
			return n
		}
		n := r.getNextID()
		idMap[id] = n
		return n
	}
	for _, n := range g.Nodes {
		id := getID(n.ID)
		nodeMap[id] = n
	}
	// initialize incomingEdges
	for _, e := range g.Edges {
		to := getID(e.To)
		incomingEdges[to] = append(incomingEdges[to], e)
	}

	var q []runNodeID
	for _, sink := range g.Sinks() {
		q = append(q, getID(sink.ID))
	}
	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		if visited[n] {
			continue
		}
		visited[n] = true
		order = append(order, n)
		for _, e := range incomingEdges[n] {
			q = append(q, getID(e.From))
		}
	}
	// reverse the order
	for i := 0; i < len(order)/2; i++ {
		j := len(order) - i - 1
		order[i], order[j] = order[j], order[i]
	}

	rs := &runState{
		nodes:        make([]runNode, len(order)),
		nodeOrder:    order,
		nodeIndexMap: make(map[runNodeID]int),
	}
	for i := range rs.nodeIndexMap {
		// initialize to -1 so that we can detect nodes that are not
		// ancestors of any sink.
		rs.nodeIndexMap[i] = -1
	}

	for i, id := range order {
		node := &rs.nodes[i]

		node.id = id

		graphNode := nodeMap[id] // BUGBUG: possible for node not to be found?
		node.node = graphNode

		var alignment GraphAlignment
		if r.g != nil {
			alignment = AlignGraphs(r.g, g)
		}
		var nodeFound bool
		if targetID, ok := alignment.NodeIdentities[graphNode.ID]; ok {
			// find the target node in previous run state, and
			// copy the UGen and value
			var tgt *runNode
			for i := range r.rs.nodes {
				n := &r.rs.nodes[i]
				if n.node.ID != targetID {
					continue
				}
				tgt = &r.rs.nodes[i]
				break
			}
			if tgt != nil {
				node.gen = tgt.gen
				node.value = tgt.value
				tgt.retained = true
				nodeFound = true
			}
		}
		if !nodeFound {
			// if a node of type out, we don't need to construct a UGen
			if graphNode.Type == "out" {
				// get the index of the out node
				idx := lang.First(graphNode.Args).(int64)
				node.gen = ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, _ []float64) {
					if int(idx) >= len(r.nextOut) {
						fmt.Printf("out of bounds output index: %d\n", idx)
						return
					}
					out := r.nextOut[int(idx)]
					clear(out)
					for _, in := range cfg.InputSamples {
						_ = in[len(out)-1]
						for i := range out {
							out[i] += in[i]
						}
					}
				})
			} else {
				node.gen = graphNode.Construct()
				if s, ok := node.gen.(ugen.Starter); ok {
					s.Start(r.ctx)
				}
			}
			node.value = make([]float64, conf.BufferSize)
		}

		predecessorNodes := map[runNodeID]struct{}{}
		incoming := incomingEdges[id]
		node.incomingEdges = incoming
		for _, e := range incoming {
			predecessorNodes[getID(e.From)] = struct{}{}
		}
		node.dependencies = predecessorNodes
		rs.nodeIndexMap[id] = i
	}
	// initialize the inputSampleMap
	for i := range rs.nodes {
		n := &rs.nodes[i]
		n.inputSampleMap = make(map[string][]float64)
		for _, e := range n.incomingEdges {
			inNode := rs.NodeByID(getID(e.From))
			n.inputSampleMap[e.Port] = inNode.value
		}
	}

	bootstrapCycles(rs)

	return rs
}

func (rs *runState) NodeByID(id runNodeID) *runNode {
	index := rs.nodeIndexMap[id]
	if index < 0 {
		return nil
	}
	return &rs.nodes[index]
}

func bootstrapCycles(rs *runState) {
	for _, info := range rs.nodes {
		if info.node.Sink {
			visited := make(map[runNodeID]struct{})
			prepCyclesDFS(rs, info.id, visited)
		}
	}
}

func prepCyclesDFS(rs *runState, nodeID runNodeID, visited map[runNodeID]struct{}) {
	visited[nodeID] = struct{}{}
	defer delete(visited, nodeID)

	info := rs.NodeByID(nodeID)
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

func (rn *runNode) run(ctx context.Context, cfg ugen.SampleConfig) {
	if rn.gen == nil {
		return
	}

	clear(rn.value)
	cfg.InputSamples = rn.inputSampleMap
	rn.gen.Gen(ctx, cfg, rn.value)
}
