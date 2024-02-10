package mrat

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/glojurelang/glojure/pkg/glj"
	"github.com/glojurelang/glojure/pkg/lang"
	value "github.com/glojurelang/glojure/pkg/lang"
	"github.com/glojurelang/glojure/pkg/runtime"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/graph"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

var (
	addedPaths = map[string]bool{}

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
)

func EvalScript(filename string) (res *graph.Graph, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v\n%s", r, debug.Stack())
		}
	}()

	require := glj.Var("glojure.core", "require")
	require.Invoke(glj.Read("mrat.core"))

	graphAtom := lang.NewAtom(glj.Read(`{:nodes [] :edges []}`))
	value.PushThreadBindings(value.NewMap(
		glj.Var("mrat.core", "*graph*"), graphAtom,
	))
	defer value.PopThreadBindings()

	// get the absolute path to the script
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}
	filename = absPath

	// get the directory of the file and the file name
	dir := filepath.Dir(filename)
	name := filepath.Base(filename)

	if !addedPaths[dir] {
		// add the directory as a fs.FS to the load path
		runtime.AddLoadPath(os.DirFS(dir))
		addedPaths[dir] = true
	}
	require.Invoke(glj.Read(strings.TrimSuffix(name, ".glj")), value.NewKeyword("reload"))

	res = &graph.Graph{BufferSize: conf.BufferSize}
	nodeMap := map[string]graph.NodeID{}

	simplifyGraph := glj.Var("mrat.core", "simplify-graph")
	g := simplifyGraph.Invoke(graphAtom.Deref())

	nodes := lang.Get(g, nodesKW)
	for s := lang.Seq(nodes); s != nil; s = lang.Next(s) {
		node := lang.First(s)
		id, nodeID, err := addNode(res, node)
		if err != nil {
			return nil, err
		}
		nodeMap[id] = nodeID
	}

	edges := lang.Get(g, edgesKW)
	for s := lang.Seq(edges); s != nil; s = lang.Next(s) {
		edge := lang.First(s)
		fromVal := lang.Get(edge, fromKW)
		from, ok := fromVal.(string)
		if !ok {
			return nil, fmt.Errorf("edge 'from' must be a string, got %T", fromVal)
		}
		toVal := lang.Get(edge, toKW)
		to, ok := toVal.(string)
		if !ok {
			return nil, fmt.Errorf("edge 'to' must be a string, got %T", toVal)
		}
		portVal := lang.Get(edge, portKW)
		port, ok := portVal.(string)
		if !ok {
			return nil, fmt.Errorf("edge 'port' must be a string, got %T", portVal)
		}

		//fmt.Printf("from: %s, to: %s, port: %s\n", from, to, port)

		res.AddEdge(nodeMap[from], nodeMap[to], port)
	}

	// fmt.Println(nodeMap)

	fmt.Println("nodes:", lang.Count(nodes))
	// fmt.Println("edges:", lang.Count(edges))
	// fmt.Println("edges:", edges)
	//	panic("done")

	return res, nil
}

func addNode(g *graph.Graph, node any) (string, graph.NodeID, error) {
	id := lang.Get(node, idKW)
	typ := lang.Get(node, typeKW)
	args := lang.Get(node, argsKW)
	ctor := lang.Get(node, ctorKW)
	sink := lang.Get(node, sinkKW)

	//fmt.Printf("id: %s, type: %s, args: %d\n", id, typ, lang.Count(args))

	var nodeID graph.NodeID
	switch typ {
	case outKW:
		outNode := g.AddOutNode(graph.WithLabel(fmt.Sprintf("out %s", lang.First(args))))
		nodeID = outNode.ID()
	default:
		var argsSlice []any
		for s := lang.Seq(args); s != nil; s = lang.Next(s) {
			argsSlice = append(argsSlice, lang.First(s))
		}
		gen, ok := lang.Apply(ctor, argsSlice).(ugen.UGen)
		if !ok {
			return "", nodeID, fmt.Errorf("no ugen provided for %s", typ)
		}
		opts := []graph.NodeOption{graph.WithLabel(fmt.Sprintf("%s", typ))}
		if s, _ := sink.(bool); s {
			opts = append(opts, graph.WithSink())
		}

		node := g.AddGeneratorNode(gen, opts...)
		nodeID = node.ID()
	}
	return fmt.Sprintf("%s", id), nodeID, nil
}
