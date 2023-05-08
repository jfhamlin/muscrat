package mrat

import (
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/value"
)

func graph() *graph.Graph {
	return nil
}

func Sin() *value.Gen {
	// nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Sin(1024)), graph.WithLabel("sin"))
	// if len(args) == 0 {
	// 	return &value.Gen{
	// 		NodeID: nodeID,
	// 	}, nil
	// }

	// freq, ok := asGen(env, args[0])
	// if !ok {
	// 	return nil, fmt.Errorf("expected generator or number as the first argument, got %v", args[0])
	// }
	// env.Graph().AddEdge(freq.NodeID, nodeID, "w")
	// args = args[1:]

	// if err := handleExtraGenArgs(env, nodeID, args); err != nil {
	// 	return nil, err
	// }

	// return &value.Gen{
	// 	NodeID: nodeID,
	// }, nil
}
