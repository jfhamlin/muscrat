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

	"github.com/jfhamlin/muscrat/pkg/graph"
	"github.com/jfhamlin/muscrat/pkg/graph2"
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
	keyKW   = lang.NewKeyword("key")
)

func EvalScript(filename string) (res *graph2.Graph, err error) {
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

	require.Invoke(glj.Read("mrat.graph"))
	simplifyGraph := glj.Var("mrat.graph", "simplify-graph")
	g := simplifyGraph.Invoke(graphAtom.Deref())
	return graph2.SExprToGraph(g), nil
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
