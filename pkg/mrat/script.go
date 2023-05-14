package mrat

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/glojurelang/glojure/glj"
	"github.com/glojurelang/glojure/value"

	"github.com/jfhamlin/muscrat/pkg/graph"
)

func EvalScript(script, filename string) (g *graph.Graph, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v\n%s", r, debug.Stack())
		}
	}()

	require := glj.Var("glojure.core", "require")
	require.Invoke(glj.Read("mrat.core"))

	g = &graph.Graph{}
	value.PushThreadBindings(value.NewMap(
		glj.Var("mrat.core", "*graph*"), g,
	))
	defer value.PopThreadBindings()

	require.Invoke(glj.Read(strings.TrimSuffix(filename, ".glj")), value.NewKeyword("reload"))

	return g, nil
}
