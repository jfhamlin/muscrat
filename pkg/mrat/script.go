package mrat

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/glojurelang/glojure/glj"
	"github.com/glojurelang/glojure/runtime"
	"github.com/glojurelang/glojure/value"

	"github.com/jfhamlin/muscrat/pkg/graph"
)

var (
	addedPaths = map[string]bool{}
)

func EvalScript(filename string) (g *graph.Graph, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v\n%s", r, debug.Stack())
		}
	}()

	require := glj.Var("glojure.core", "require")
	require.Invoke(glj.Read("mrat.core"))

	g = &graph.Graph{BufferSize: bufferSize}
	value.PushThreadBindings(value.NewMap(
		glj.Var("mrat.core", "*graph*"), g,
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

	return g, nil
}
