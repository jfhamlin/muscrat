package value

import (
	"context"
	"io"

	"github.com/jfhamlin/muscrat/internal/pkg/graph"
)

// Environment is an interface for execution environments.
type Environment interface {
	// PushScope returns a new Environment with a scope nested inside
	// this environment's scope.
	PushScope() Environment

	// Define defines a variable in the current scope.
	Define(name string, v Value)

	// Eval evaluates a value representing an expression in this
	// environment.
	Eval(expr Value) (Value, error)

	// ResolveFile looks up a file in the environment. It should expand
	// relative paths to absolute paths. Relative paths are searched for
	// in the environments load paths.
	ResolveFile(path string) (string, bool)

	// PushLoadPaths adds paths to the environment's list of load
	// paths. The provided paths will be searched for relative paths
	// first in the returned environment.
	PushLoadPaths(paths []string) Environment

	// Graph returns the graph associated with this environment.
	Graph() *graph.Graph

	// Stdout returns the standard output stream for this environment.
	Stdout() io.Writer

	// Context returns the context associated with this environment.
	Context() context.Context
}
