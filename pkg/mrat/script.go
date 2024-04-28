package mrat

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/glojurelang/glojure/pkg/glj"
	"github.com/glojurelang/glojure/pkg/lang"
	value "github.com/glojurelang/glojure/pkg/lang"
	"github.com/glojurelang/glojure/pkg/runtime"

	"github.com/jfhamlin/muscrat/pkg/console"
	"github.com/jfhamlin/muscrat/pkg/graph"
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

type (
	consoleWriter struct {
		sb strings.Builder
	}

	UGenArg struct {
		Name    string `json:"name"`
		Default any    `json:"default"`
		Doc     string `json:"doc"`
	}

	Symbol struct {
		Name     string    `json:"name"`
		Group    string    `json:"group"`
		Doc      string    `json:"doc"`
		Arglists []any     `json:"arglists"`
		UGenArgs []UGenArg `json:"ugenargs"`
	}
)

func (cw *consoleWriter) Write(p []byte) (n int, err error) {
	// write each line to the console
	for _, char := range p {
		if char == '\n' {
			console.Log(console.Info, cw.sb.String(), nil)
			cw.sb.Reset()
		} else {
			cw.sb.WriteByte(char)
		}
	}
	return len(p), nil
}

func EvalScript(filename string) (res *graph.Graph, err error) {
	console.Log(console.Info, fmt.Sprintf("evaluating %s", filename), nil)
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
		glj.Var("glojure.core", "*out*"), &consoleWriter{},
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
	return graph.SExprToGraph(g), nil
}

func GetNSPublics() []Symbol {
	require := glj.Var("glojure.core", "require")
	require.Invoke(glj.Read("mrat.core"))

	nsPublics := glj.Var("glojure.core", "ns-publics")
	publics := nsPublics.Invoke(glj.Read("mrat.core"))

	docgroupKW := lang.NewKeyword("docgroup")
	docKW := lang.NewKeyword("doc")
	argsKW := lang.NewKeyword("arglists")
	ugenargsKW := lang.NewKeyword("ugenargs")
	nameKW := lang.NewKeyword("name")
	defaultKW := lang.NewKeyword("default")

	var res []Symbol
	for s := lang.Seq(publics); s != nil; s = s.Next() {
		kv := s.First().(*lang.MapEntry)
		name := kv.Key().(*lang.Symbol).Name()
		val := kv.Val().(*lang.Var)
		meta := val.Meta()

		docgroup, _ := meta.ValAt(docgroupKW).(string)
		doc, _ := meta.ValAt(docKW).(string)
		arglists := meta.ValAt(argsKW)
		var arglist []any
		for s := lang.Seq(arglists); s != nil; s = s.Next() {
			arglist = append(arglist, s.First())
		}
		uargs := meta.ValAt(ugenargsKW)
		var ugenargs []UGenArg
		for s := lang.Seq(uargs); s != nil; s = s.Next() {
			m := s.First().(*lang.Map)
			name := m.ValAt(nameKW).(string)
			defaultVal := m.ValAt(defaultKW)
			doc, _ := m.ValAt(docKW).(string)
			ugenargs = append(ugenargs, UGenArg{
				Name:    name,
				Default: defaultVal,
				Doc:     doc,
			})
		}

		sym := Symbol{
			Name:     name,
			Group:    docgroup,
			Doc:      doc,
			Arglists: arglist,
			UGenArgs: ugenargs,
		}
		res = append(res, sym)
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Group < res[j].Group {
			return true
		}
		if res[i].Group > res[j].Group {
			return false
		}
		return res[i].Name < res[j].Name
	})

	return res
}
