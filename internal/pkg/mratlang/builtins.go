package mratlang

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jfhamlin/muscrat/internal/pkg/generator"
	"github.com/jfhamlin/muscrat/internal/pkg/graph"
	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/value"
	"github.com/jfhamlin/muscrat/internal/pkg/wavtabs"
	"github.com/jfhamlin/muscrat/pkg/freeverb"
)

var (
	builtinPackages []*Package
)

func init() {
	builtinPackages = []*Package{
		&Package{
			Name: "mrat.core",
			Symbols: []*Symbol{
				// importing/requiring other packages
				funcSymbol("load-file", loadFileBuiltin),
				// list functions
				funcSymbol("list", listBuiltin),
				funcSymbol("length", lengthBuiltin),
				funcSymbol("concat", concatBuiltin),
				funcSymbol("first", firstBuiltin),
				funcSymbol("rest", restBuiltin),
				// math functions
				funcSymbol("pow", powBuiltin),
				funcSymbol("*", mulBuiltin),
				funcSymbol("/", divBuiltin),
				funcSymbol("+", addBuiltin),
				funcSymbol("-", subBuiltin),
				funcSymbol("<", ltBuiltin),
				// function application
				funcSymbol("apply", applyBuiltin),
				// test predicates
				funcSymbol("eq?", eqBuiltin),
				funcSymbol("empty?", emptyBuiltin),
				funcSymbol("not-empty?", notEmptyBuiltin),
				// boolean functions
				funcSymbol("not", notBuiltin),
				// plumbing
				funcSymbol("~pipe", pipeBuiltin),
				funcSymbol("pipeset", pipesetBuiltin),
			},
		},
		&Package{
			Name: "mrat.core.io",
			Symbols: []*Symbol{
				funcSymbol("print", printBuiltin),
				funcSymbol("out", outBuiltin),
			},
		},
		&Package{
			Name: "mrat.math.rand",
			Symbols: []*Symbol{
				funcSymbol("trand", trandBuiltin),
			},
		},
		&Package{
			Name: "mrat.osc",
			Symbols: []*Symbol{
				funcSymbol("sin", sinBuiltin),
				funcSymbol("saw", sawBuiltin),
				funcSymbol("sqr", sqrBuiltin),
				funcSymbol("tri", triBuiltin),
				funcSymbol("noise", noiseBuiltin),
			},
		},
		&Package{
			Name: "mrat.effects",
			Symbols: []*Symbol{
				funcSymbol("freeverb", freeverbBuiltin),
				funcSymbol("clip", clipBuiltin),
				funcSymbol("delay", delayBuiltin),
				funcSymbol("mixer", mixerBuiltin),
				funcSymbol("env", envBuiltin),
				funcSymbol("~lores", loresBuiltin),
			},
		},
	}
}

func addBuiltins(env *environment) {
	for _, pkg := range builtinPackages {
		for _, sym := range pkg.Symbols {
			name := pkg.Name + "." + sym.Name
			if pkg.Name == "mrat.core" {
				// core symbols are available in the global namespace.
				name = sym.Name
			}
			env.Define(name, sym.Value)
		}
	}
}

func funcSymbol(name string, fn func(value.Environment, []value.Value) (value.Value, error)) *Symbol {
	return &Symbol{
		Name: name,
		Value: &value.BuiltinFunc{
			Applyer: value.ApplyerFunc(fn),
			Name:    name,
		},
	}
}

func loadFile(env value.Environment, filename string) error {
	absFile, ok := env.ResolveFile(filename)
	if !ok {
		return fmt.Errorf("could not resolve file %v", filename)
	}

	buf, err := ioutil.ReadFile(absFile)
	if err != nil {
		return fmt.Errorf("error reading file %v: %w", filename, err)
	}

	prog, err := Parse(strings.NewReader(string(buf)), WithFilename(absFile))
	if err != nil {
		return fmt.Errorf("error parsing file %v: %w", filename, err)
	}

	loadEnv := env.PushLoadPaths([]string{filepath.Dir(absFile)})
	_, _, err = prog.Eval(withEnv(loadEnv))
	if err != nil {
		return fmt.Errorf("error evaluating file %v: %w", filename, err)
	}

	return nil
}

func loadFileBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("load-file expects 1 argument, got %v", len(args))
	}
	filename, ok := args[0].(*value.Str)
	if !ok {
		return nil, fmt.Errorf("load-file expects a string, got %v", args[0])
	}
	return nil, loadFile(env, filename.Value)
}

func listBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	return value.NewList(args), nil
}

func lengthBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("length expects 1 argument, got %v", len(args))
	}
	switch arg := args[0].(type) {
	case *value.List:
		return value.NewNum(float64(len(arg.Items))), nil
	default:
		return nil, fmt.Errorf("length expects a list, got %v", arg)
	}
}

func concatBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	var res []value.Value
	for _, arg := range args {
		switch arg := arg.(type) {
		case *value.List:
			res = append(res, arg.Items...)
		default:
			return nil, fmt.Errorf("invalid type for concat: %v", arg)
		}
	}
	return value.NewList(res), nil
}

func firstBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("first expects 1 argument, got %v", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return nil, fmt.Errorf("first expects a list, got %v", args[0])
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("first expects a non-empty list, got %v", list)
	}
	return list.Items[0], nil
}

func restBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rest expects 1 argument, got %v", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return nil, fmt.Errorf("rest expects a list, got %v", args[0])
	}
	if len(list.Items) == 0 {
		return list, nil
	}
	return value.NewList(list.Items[1:]), nil
}

func notBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("not expects 1 argument, got %v", len(args))
	}
	switch arg := args[0].(type) {
	case *value.Bool:
		return value.NewBool(!arg.Value), nil
	default:
		return nil, fmt.Errorf("not expects a boolean, got %v", arg)
	}
}

func eqBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("eq? expects 2 arguments, got %v", len(args))
	}
	return value.NewBool(args[0].Equal(args[1])), nil
}

func emptyBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("empty? expects 1 argument, got %v", len(args))
	}
	list, ok := args[0].(*value.List)
	if !ok {
		return nil, fmt.Errorf("empty? expects a list, got %v", args[0])
	}
	return value.NewBool(len(list.Items) == 0), nil
}

func notEmptyBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	v, err := emptyBuiltin(env, args)
	if err != nil {
		return nil, err
	}
	return notBuiltin(env, []value.Value{v})
}

func powBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("pow expects 2 arguments, got %v", len(args))
	}
	a, ok := args[0].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("pow expects a number, got %v", args[0])
	}
	b, ok := args[1].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("pow expects a number, got %v", args[1])
	}
	return value.NewNum(math.Pow(a.Value, b.Value)), nil
}

func mulBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	var coeff float64 = 1
	gens := []*value.Gen{}

	// multiply all number arguments together
	for _, arg := range args {
		switch arg := arg.(type) {
		case *value.Num:
			coeff *= arg.Value
		case *value.Gen:
			gens = append(gens, arg)
		default:
			return nil, fmt.Errorf("invalid type for *: %v", arg)
		}
	}
	// if there are no generators, return the result of multiplying all the numbers
	if len(gens) == 0 {
		return value.NewNum(coeff), nil
	}

	// otherwise, create a new constant generator node for the coefficient
	if coeff != 1 {
		constNodeID := env.Graph().AddGeneratorNode(generator.NewConstant(coeff), graph.WithLabel(fmt.Sprintf("%v", coeff)))
		gens = append(gens, &value.Gen{NodeID: constNodeID})
	}

	// create a new generator node for the product
	nodeID := env.Graph().AddGeneratorNode(generator.NewProduct(), graph.WithLabel("*"))
	for i, gen := range gens {
		env.Graph().AddEdge(gen.NodeID, nodeID, fmt.Sprintf("$%d", i))
	}
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func divBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("div expects 2 arguments, got %v", len(args))
	}
	num, ok := args[0].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("div expects a number as the first argument, got %v", args[0])
	}
	denom, ok := args[1].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("div expects a number as the second argument, got %v", args[1])
	}
	// TODO: handle generators
	return value.NewNum(num.Value / denom.Value), nil
}

func addBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	var sum float64 = 0
	gens := []*value.Gen{}

	// sum all number arguments together
	for _, arg := range args {
		switch arg := arg.(type) {
		case *value.Num:
			sum += arg.Value
		case *value.Gen:
			gens = append(gens, arg)
		default:
			return nil, fmt.Errorf("invalid type for +: %v", arg)
		}
	}
	// if there are no generators, return the result of summing all the numbers
	if len(gens) == 0 {
		return value.NewNum(sum), nil
	}

	// otherwise, if the sum is not zero, create a new constant
	// generator node for the sum
	if sum != 0 {
		constNodeID := env.Graph().AddGeneratorNode(generator.NewConstant(sum), graph.WithLabel(fmt.Sprintf("%v", sum)))
		gens = append(gens, &value.Gen{NodeID: constNodeID})
	}

	// create a new generator node for the sum
	nodeID := env.Graph().AddGeneratorNode(generator.NewSum(), graph.WithLabel("+"))
	for i, gen := range gens {
		env.Graph().AddEdge(gen.NodeID, nodeID, fmt.Sprintf("$%d", i))
	}
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func subBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("sub expects 2 arguments, got %v", len(args))
	}
	a, ok := args[0].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("sub expects a number as the first argument, got %v", args[0])
	}
	b, ok := args[1].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("sub expects a number as the second argument, got %v", args[1])
	}

	// TODO: handle generators
	return value.NewNum(a.Value - b.Value), nil
}

func ltBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("< expects 2 arguments, got %v", len(args))
	}
	a, ok := args[0].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("< expects a number as the first argument, got %v", args[0])
	}
	b, ok := args[1].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("< expects a number as the second argument, got %v", args[1])
	}

	return value.NewBool(a.Value < b.Value), nil
}

type pipe struct {
	inputSet bool
}

func (p *pipe) GenerateSamples(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
	if in, ok := cfg.InputSamples["in"]; ok {
		return in
	}
	return make([]float64, n)
}

func pipeBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("pipe expects 0 arguments, got %v", len(args))
	}
	nodeID := env.Graph().AddGeneratorNode(&pipe{}, graph.WithLabel("~pipe"))
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func pipesetBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("pipeset expects 2 arguments, got %v", len(args))
	}

	// get the generator
	gen, ok := args[0].(*value.Gen)
	if !ok {
		return nil, fmt.Errorf("pipeset expects a pipe generator as the first argument, got %v", args[0])
	}
	node := env.Graph().Node(gen.NodeID)
	if node == nil {
		return nil, fmt.Errorf("invalid generator node: %v", gen.NodeID)
	}
	genNode, ok := node.(*graph.GeneratorNode)
	if !ok {
		return nil, fmt.Errorf("pipeset expects a pipe generator as the first argument, got %v", args[0])
	}
	pipeGen, ok := genNode.Generator.(*pipe)
	if !ok {
		return nil, fmt.Errorf("pipeset expects a pipe generator as the first argument, got %v", args[0])
	}

	if pipeGen.inputSet {
		return nil, fmt.Errorf("pipeset called twice on the same pipe")
	}
	pipeGen.inputSet = true

	input, ok := asGen(env, args[1])
	if !ok {
		return nil, fmt.Errorf("pipeset expects a generator as the second argument, got %v", args[1])
	}

	// add an edge from the input generator to the pipe generator
	env.Graph().AddEdge(input.NodeID, gen.NodeID, "in")
	return nil, nil
}

func applyBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("apply expects 2 arguments, got %v", len(args))
	}
	// the first argument should be an applyer, the second a list
	applyer, ok := args[0].(value.Applyer)
	if !ok {
		return nil, fmt.Errorf("apply expects a function as the first argument, got %v", args[0])
	}
	list, ok := args[1].(*value.List)
	if !ok {
		return nil, fmt.Errorf("apply expects a list as the second argument, got %v", args[1])
	}
	return applyer.Apply(env, list.Items)
}

func printBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	for i, arg := range args {
		if arg == nil {
			// TODO: add nil to the type system
			env.Stdout().Write([]byte("nil"))
		} else {
			str, ok := arg.(*value.Str)
			if !ok {
				env.Stdout().Write([]byte(arg.String()))
			} else {
				env.Stdout().Write([]byte(str.Value))
			}
		}
		if i < len(args)-1 {
			env.Stdout().Write([]byte(" "))
		}
	}
	env.Stdout().Write([]byte("\n"))
	return nil, nil
}

func linRand(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func expRand(min, max float64) float64 {
	return min * math.Pow(max/min, rand.Float64())
}

func NewTrandGenerator(pick func(min, max float64) float64) generator.SampleGenerator {
	triggered := false
	value := make([]float64, 0, 1)
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		trigger := cfg.InputSamples["trigger"]
		min := cfg.InputSamples["min"]
		max := cfg.InputSamples["max"]

		if len(value) == 0 {
			value = append(value, pick(min[0], max[0]))
		}

		res := make([]float64, n)
		for i := 0; i < n; i++ {
			if !triggered && trigger[i] > 0 {
				triggered = true
				value[0] = pick(min[i], max[i])
			}
			if triggered && trigger[i] <= 0 {
				triggered = false
			}

			res[i] = value[0]
		}
		return res
	})
}

func trandBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	// trand takes three required arguments: a generator that triggers
	// random number selection and two generators that produce the min
	// and max values for the random number selection.
	//
	// trand also takes an optional argument that specifies the random
	// number selection function. The default is linear.
	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("trand expects 3 or 4 arguments, got %v", len(args))
	}
	gens := make([]*value.Gen, 3)
	for i, arg := range args[:3] {
		gen, ok := asGen(env, arg)
		if !ok {
			return nil, fmt.Errorf("trand expects generators as arguments, got %v", arg)
		}
		gens[i] = gen
	}
	randFn := linRand
	if len(args) == 4 {
		kw, ok := args[3].(*value.Keyword)
		if !ok {
			return nil, fmt.Errorf("trand expects a keyword as the fourth argument, got %v", args[3])
		}
		switch kw.Value {
		case "lin":
			randFn = linRand
		case "exp":
			randFn = expRand
		default:
			return nil, fmt.Errorf("trand does not recognize the keyword %v", kw.Value)
		}
	}

	nodeID := env.Graph().AddGeneratorNode(NewTrandGenerator(randFn), graph.WithLabel("trand"))
	env.Graph().AddEdge(gens[0].NodeID, nodeID, "trigger")
	env.Graph().AddEdge(gens[1].NodeID, nodeID, "min")
	env.Graph().AddEdge(gens[2].NodeID, nodeID, "max")
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func sinBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	var freq graph.NodeID
	if len(args) == 0 {
		freq = env.Graph().AddGeneratorNode(generator.NewConstant(440), graph.WithLabel("440"))
	} else {
		switch arg := args[0].(type) {
		case *value.Num:
			freq = env.Graph().AddGeneratorNode(generator.NewConstant(arg.Value), graph.WithLabel(fmt.Sprintf("%v", arg.Value)))
		case *value.Gen:
			freq = arg.NodeID
		default:
			return nil, fmt.Errorf("invalid type for sin frequency: %v", arg)
		}
	}
	nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Sin(1024)), graph.WithLabel("sin"))
	env.Graph().AddEdge(freq, nodeID, "w")
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func sawBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	var freq graph.NodeID
	if len(args) == 0 {
		freq = env.Graph().AddGeneratorNode(generator.NewConstant(440), graph.WithLabel("440"))
	} else {
		switch arg := args[0].(type) {
		case *value.Num:
			freq = env.Graph().AddGeneratorNode(generator.NewConstant(arg.Value), graph.WithLabel(fmt.Sprintf("%v", arg.Value)))
		case *value.Gen:
			freq = arg.NodeID
		default:
			return nil, fmt.Errorf("invalid type for saw frequency: %v", arg)
		}
	}
	nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Saw(1024)), graph.WithLabel("saw"))
	env.Graph().AddEdge(freq, nodeID, "w")
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func triBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	var freq graph.NodeID
	if len(args) == 0 {
		freq = env.Graph().AddGeneratorNode(generator.NewConstant(440), graph.WithLabel("440"))
	} else {
		switch arg := args[0].(type) {
		case *value.Num:
			freq = env.Graph().AddGeneratorNode(generator.NewConstant(arg.Value), graph.WithLabel(fmt.Sprintf("%v", arg.Value)))
		case *value.Gen:
			freq = arg.NodeID
		default:
			return nil, fmt.Errorf("invalid type for tri frequency: %v", arg)
		}
	}
	nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Tri(1024)), graph.WithLabel("tri"))
	env.Graph().AddEdge(freq, nodeID, "w")
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func squareWaveSample(dutyCycle, phase float64) float64 {
	phase = phase - math.Floor(phase)
	if phase < dutyCycle {
		return 1
	}
	return -1
}

func NewSquareGenerator() generator.SampleGenerator {
	phase := 0.0
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		dcs := cfg.InputSamples["dc"]
		ws := cfg.InputSamples["w"]
		res := make([]float64, n)
		for i := 0; i < n; i++ {
			w := 440.0
			dc := 0.5
			if i < len(ws) {
				w = ws[i]
			}
			if i < len(dcs) {
				dc = dcs[i]
			}
			res[i] = squareWaveSample(dc, phase)
			phase += w / float64(cfg.SampleRateHz)
			phase -= math.Floor(phase)
		}
		return res
	})
}

func sqrBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	var freq graph.NodeID
	if len(args) == 0 {
		freq = env.Graph().AddGeneratorNode(generator.NewConstant(440), graph.WithLabel("440"))
	} else {
		gen, ok := asGen(env, args[0])
		if !ok {
			return nil, fmt.Errorf("invalid type for sin frequency: %v", args[0])
		}
		freq = gen.NodeID
		args = args[1:]
	}
	var dutyCycle graph.NodeID
	if len(args) == 0 {
		dutyCycle = env.Graph().AddGeneratorNode(generator.NewConstant(0.5), graph.WithLabel("0.5"))
	} else {
		gen, ok := asGen(env, args[0])
		if !ok {
			return nil, fmt.Errorf("invalid type for saw duty cycle: %v", args[0])
		}
		dutyCycle = gen.NodeID
	}

	nodeID := env.Graph().AddGeneratorNode(NewSquareGenerator(), graph.WithLabel("sqr"))
	env.Graph().AddEdge(freq, nodeID, "w")
	env.Graph().AddEdge(dutyCycle, nodeID, "dc")
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func outBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	sinks := env.Graph().Sinks()
	if len(sinks) == 0 {
		sinks = append(sinks, env.Graph().AddSinkNode(graph.WithLabel("out")))
	}
	for _, arg := range args {
		switch arg := arg.(type) {
		case *value.Gen:
			for _, sink := range sinks {
				env.Graph().AddEdge(arg.NodeID, sink.ID(), arg.String())
			}
		default:
			return nil, fmt.Errorf("invalid type for out: %v", arg)
		}
	}
	return nil, nil
}

func NewFreeverbGenerator() generator.SampleGenerator {
	revmod := freeverb.NewRevModel()
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		if wet := cfg.InputSamples["wet"]; len(wet) > 0 {
			revmod.SetWet(float32(wet[0]))
		}
		if damp := cfg.InputSamples["damp"]; len(damp) > 0 {
			revmod.SetDamp(float32(damp[0]))
		}
		if room := cfg.InputSamples["room"]; len(room) > 0 {
			revmod.SetRoomSize(float32(room[0]))
		}

		input32 := make([]float32, n)
		for i := 0; i < n; i++ {
			input32[i] = float32(cfg.InputSamples["$0"][i])
		}
		outputLeft := make([]float32, n)
		outputRight := make([]float32, n)
		revmod.ProcessReplace(input32, input32, outputLeft, outputRight, n, 1)
		output := make([]float64, n)
		for i := 0; i < n; i++ {
			output[i] = 0.5 * (float64(outputLeft[i]) + float64(outputRight[i]))
		}

		return output
	})
}

func freeverbBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	// take a single argument, which should be a Gen
	if len(args) != 1 {
		return nil, fmt.Errorf("freeverb expects 1 argument, got %v", len(args))
	}
	gen, ok := args[0].(*value.Gen)
	if !ok {
		return nil, fmt.Errorf("freeverb expects a Gen as the first argument, got %v", args[0])
	}
	nodeID := env.Graph().AddGeneratorNode(NewFreeverbGenerator(), graph.WithLabel("freeverb"))
	env.Graph().AddEdge(gen.NodeID, nodeID, "$0")
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func NewClipGenerator(min, max float64) generator.SampleGenerator {
	// TODO: min and max should be inputs
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		input := cfg.InputSamples["$0"]
		output := make([]float64, n)
		for i := 0; i < n; i++ {
			output[i] = math.Max(min, math.Min(max, input[i]))
		}
		return output
	})
}

func clipBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	// clip takes three arguments: a Gen, a min, and a max
	if len(args) != 3 {
		return nil, fmt.Errorf("clip expects 3 arguments, got %v", len(args))
	}
	gen, ok := args[0].(*value.Gen)
	if !ok {
		return nil, fmt.Errorf("clip expects a Gen as the first argument, got %v", args[0])
	}
	min, ok := args[1].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("clip expects a Num as the second argument, got %v", args[1])
	}
	max, ok := args[2].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("clip expects a Num as the third argument, got %v", args[2])
	}
	nodeID := env.Graph().AddGeneratorNode(NewClipGenerator(min.Value, max.Value), graph.WithLabel("clip"))
	env.Graph().AddEdge(gen.NodeID, nodeID, "$0")
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func NewDelayGenerator() generator.SampleGenerator {
	// Simulate a tape delay by using a buffer of samples with a read
	// and write pointer. If the delay is changed, we simulate a
	// physical read/write head by maintaining a sample velocity for the
	// read head. The write head is always at the end of the buffer. The
	// read head can never move backwards, so if the delay is decreased,
	// the read head will accelerate, and if the delay is increased, the
	// read head will decelerate.
	var tape []float64
	var readHead float64
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		in := cfg.InputSamples["$0"]

		// var targetDelaySecs float64
		// var targetDelaySamps float64
		// var actualDelaySamps float64
		for i := 0; i < n; i++ {
			delaySeconds := cfg.InputSamples["delay"][i]
			if delaySeconds < 0 {
				delaySeconds = 0
			}
			delaySamples := delaySeconds * float64(cfg.SampleRateHz)
			// handle the initialization case, where the tape hasn't been set up yet.
			if tape == nil {
				tape = make([]float64, int(delaySeconds*float64(cfg.SampleRateHz)))
			}
			actualDelaySamples := float64(len(tape)) - readHead

			tape = append(tape, in[i])

			if len(tape) == 1 {
				res[i] = tape[0]
			} else {
				// read the sample from the tape at the read head with linear interpolation
				// between the two adjacent samples.
				readHeadInt := int(readHead)
				readHeadFrac := readHead - float64(readHeadInt)
				res[i] = tape[readHeadInt]*(1-readHeadFrac) + tape[readHeadInt+1]*readHeadFrac
			}

			const maxStep = 2
			const minStep = 1 / maxStep

			// update the read head position with max and min bounds to prevent
			// the read head from moving backwards or infinitely forward.
			if delaySamples == 0 && actualDelaySamples > 0 {
				readHead += maxStep
			} else if actualDelaySamples > maxStep*delaySamples {
				readHead += maxStep
			} else if actualDelaySamples < minStep*delaySamples {
				readHead += minStep
			} else {
				vel := actualDelaySamples / delaySamples
				if math.IsNaN(vel) {
					readHead += maxStep
				} else {
					readHead += math.Max(minStep, math.Min(maxStep, vel))
				}
			}
			if readHead >= float64(len(tape)) {
				readHead = 0
				tape = tape[:0]
			}
			// drop samples that have already been read from the tape.
			if readHead > 1 {
				tape = tape[int(readHead):]
				readHead = readHead - math.Floor(readHead)
			}

			// targetDelaySecs = delaySeconds
			// targetDelaySamps = delaySamples
			// actualDelaySamps = actualDelaySamples
		}

		// fmt.Printf("sample diff: %v, target delay sec: %v, target delay samps: %v, actual delay samps: %v, read head: %v\n, ratio: %v\n", targetDelaySamps-actualDelaySamps, targetDelaySecs, targetDelaySamps, actualDelaySamps, readHead, actualDelaySamps/targetDelaySamps)

		return res
	})
}

func delayBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	// delay takes two arguments that can be converted to generators.
	// it then creates a delay generator, sending the first argument to
	// $0 and the second argument to "delay".
	if len(args) != 2 {
		return nil, fmt.Errorf("delay expects 2 arguments, got %v", len(args))
	}
	gen, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("delay expects a Gen as the first argument, got %v", args[0])
	}
	delay, ok := asGen(env, args[1])
	if !ok {
		return nil, fmt.Errorf("delay expects a Gen as the second argument, got %v", args[1])
	}
	nodeID := env.Graph().AddGeneratorNode(NewDelayGenerator(), graph.WithLabel("delay"))
	env.Graph().AddEdge(gen.NodeID, nodeID, "$0")
	env.Graph().AddEdge(delay.NodeID, nodeID, "delay")
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func NewMixerGenerator() generator.SampleGenerator {
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := 0; i < n; i++ {
			var sum float64
			for _, in := range cfg.InputSamples {
				sum += in[i]
			}
			res[i] = sum / float64(len(cfg.InputSamples))
		}
		return res
	})
}

func mixerBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	// mixer takes a variable number of arguments that can be converted to generators
	// and mixes them together.
	gens := make([]*value.Gen, len(args))
	for i, arg := range args {
		gen, ok := asGen(env, arg)
		if !ok {
			return nil, fmt.Errorf("mixer expects a Gen as the %v argument, got %v", i, arg)
		}
		gens[i] = gen
	}
	nodeID := env.Graph().AddGeneratorNode(NewMixerGenerator(), graph.WithLabel("mixer"))
	for i, gen := range gens {
		env.Graph().AddEdge(gen.NodeID, nodeID, fmt.Sprintf("$%v", i))
	}
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

var Noise = generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
	res := make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = 2*rand.Float64() - 1
	}
	return res
})

func noiseBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("noise expects 0 arguments, got %v", len(args))
	}
	nodeID := env.Graph().AddGeneratorNode(Noise, graph.WithLabel("noise"))
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func getSampleArrays(inputs map[string][]float64, name string) [][]float64 {
	var numSamples int

	prefix := name + "$"
	var res [][]float64
	for key, in := range inputs {
		if strings.HasPrefix(key, prefix) {
			idx, err := strconv.Atoi(key[len(prefix):])
			if err != nil {
				continue
			}
			for idx >= len(res) {
				res = append(res, nil)
			}
			res[idx] = in
			if numSamples == 0 {
				numSamples = len(in)
			}
		}
	}
	// fill in any missing inputs with arrays of zeros.
	for i, in := range res {
		if len(in) == 0 {
			res[i] = make([]float64, numSamples)
		}
	}
	return res
}

func NewEnvelopeGenerator(interpolation string) generator.SampleGenerator {
	// Behavior follows that of SuperCollider's Env/EnvGen
	// https://doc.sccode.org/Classes/Env.html

	triggered := false
	triggerTime := 0.0
	lastGate := false
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)

		levels := getSampleArrays(cfg.InputSamples, "level")
		times := getSampleArrays(cfg.InputSamples, "time")
		gate := cfg.InputSamples["trigger"]
		if len(gate) == 0 {
			gate = make([]float64, n)
		}

		for i := 0; i < n; i++ {
			var envDur float64
			for _, t := range times {
				envDur += t[i]
			}

			if (!triggered || triggerTime > envDur) && gate[i] > 0 && !lastGate {
				triggered = true
				triggerTime = 0
			}

			lastGate = gate[i] > 0

			if !triggered {
				res[i] = levels[0][i]
				continue
			}
			if triggerTime > envDur {
				res[i] = levels[len(levels)-1][i]
				continue
			}

			// interpolate between the two levels adjacent to the current
			// time.
			var timeSum float64
			for j, t := range times {
				timeSum += t[i]
				if timeSum >= triggerTime {
					level1 := levels[j][i]
					level2 := levels[j+1][i]
					// interpolate between levels[j] and levels[j+1]
					// at time triggerTime
					switch interpolation {
					case "lin":
						res[i] = level1 + (level2-level1)*(triggerTime-(timeSum-t[i]))/t[i]
					case "exp":
						res[i] = level1 * math.Pow(level2/level1, (triggerTime-(timeSum-t[i]))/t[i])
					}
					break
				}
			}
			triggerTime += 1 / float64(cfg.SampleRateHz)
		}
		return res
	})
}

func envBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	// env takes three arguments:
	//
	// 1. a gating signal. when the signal is > 0, the envelope is triggered.
	//
	// 2. a list of levels, which are the values of the envelope at each
	// point. these must be convertible to Gens.
	//
	// 3. a list of durations, which are the durations of each segment
	// of the envelope. these must be convertible to Gens.

	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("env expects 3 or 4 arguments, got %v", len(args))
	}

	// the first argument is the trigger signal.
	trigger, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("env expects a Gen as the first argument, got %v", args[0])
	}

	// the second argument is the list of levels.
	levels, ok := asList(args[1])
	if !ok {
		return nil, fmt.Errorf("env expects a list as the second argument, got %v", args[1])
	}
	levelGens := make([]*value.Gen, len(levels))
	for i, level := range levels {
		gen, ok := asGen(env, level)
		if !ok {
			return nil, fmt.Errorf("env expects a Gen as the %vth element of the second argument, got %v", i, level)
		}
		levelGens[i] = gen
	}

	// the third argument is the list of durations.
	durations, ok := asList(args[2])
	if !ok {
		return nil, fmt.Errorf("env expects a list as the third argument, got %v", args[2])
	}
	durationGens := make([]*value.Gen, len(durations))
	for i, duration := range durations {
		gen, ok := asGen(env, duration)
		if !ok {
			return nil, fmt.Errorf("env expects a Gen as the %vth element of the third argument, got %v", i, duration)
		}
		durationGens[i] = gen
	}

	if len(levelGens) != len(durationGens)+1 {
		return nil, fmt.Errorf("env expects the number of levels to be one more than the number of durations, got %v levels and %v durations", len(levelGens), len(durationGens))
	}

	// the optional fourth argument is the type of interpolation to use.
	interpolation := "lin"
	if len(args) == 4 {
		interpKey, ok := args[3].(*value.Keyword)
		if !ok {
			return nil, fmt.Errorf("env expects a keyword as the fourth argument, got %v", args[3])
		}
		interpolation = interpKey.Value
		if interpolation != "lin" && interpolation != "exp" {
			return nil, fmt.Errorf("env expects the fourth argument to be either :lin or :exp, got %v", interpolation)
		}
	}

	// create the envelope generator.
	nodeID := env.Graph().AddGeneratorNode(NewEnvelopeGenerator(interpolation), graph.WithLabel("env"))
	// nodeID := env.Graph().AddGeneratorNode(generator.NewConstant(1), graph.WithLabel("env"))
	env.Graph().AddEdge(trigger.NodeID, nodeID, "trigger")
	for i, gen := range levelGens {
		env.Graph().AddEdge(gen.NodeID, nodeID, fmt.Sprintf("level$%v", i))
	}
	for i, gen := range durationGens {
		env.Graph().AddEdge(gen.NodeID, nodeID, fmt.Sprintf("time$%v", i))
	}

	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func NewLowpassFilterGenerator() generator.SampleGenerator {
	// Translated from the SuperCollider extension source code here, which in turn mimics the
	// max/msp lores~ object:
	// https://github.com/v7b1/vb_UGens/blob/fea1587dd2165457c4a016214d17216987b56f00/projects/vbUtils/vbUtils.cpp
	var a1, a2, fqterm, resterm, scale, ym1, ym2 float64
	lastCut, lastRes := -1.0, -1.0
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		in := cfg.InputSamples["in"]
		cuts := cfg.InputSamples["cutoff"]
		ress := cfg.InputSamples["resonance"]
		out := make([]float64, n)

		for i := 0; i < n; i++ {
			cut := cuts[i]
			res := ress[i]
			// clamp resonance to [0, 1)
			if res < 0 {
				res = 0
			} else if res >= 1 {
				res = 1.0 - 1e-20
			}

			if cut != lastCut || res != lastRes {
				if res != lastRes {
					resterm = math.Exp(res*0.125) * 0.882497
				}
				if cut != lastCut {
					fqterm = math.Cos(cut * math.Pi * 2 / float64(cfg.SampleRateHz))
				}
				// recalculate the coefficients.
				a1 = -2 * resterm * fqterm
				a2 = resterm * resterm
				scale = 1 + a1 + a2
				lastCut, lastRes = cut, res
			}
			val := in[i]
			temp := ym1
			ym1 = scale*val - a1*ym1 - a2*ym2
			ym2 = temp
			out[i] = ym1
		}
		return out
	})
}

func loresBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	// lores takes three arguments:
	//
	// 1. a signal to be low-pass filtered.
	//
	// 2. a cutoff frequency, in Hz.
	//
	// 3. a resonance value, from 0 to 1.

	if len(args) != 3 {
		return nil, fmt.Errorf("~lores expects 3 arguments, got %v", len(args))
	}

	gen, ok := args[0].(*value.Gen)
	if !ok {
		return nil, fmt.Errorf("~lores expects a Gen as the first argument, got %v", args[0])
	}

	cutoff, ok := asGen(env, args[1])
	if !ok {
		return nil, fmt.Errorf("~lores expects a gennable as the second argument, got %v", args[1])
	}
	resonance, ok := asGen(env, args[2])
	if !ok {
		return nil, fmt.Errorf("~lores expects a gennable as the third argument, got %v", args[2])
	}

	nodeID := env.Graph().AddGeneratorNode(NewLowpassFilterGenerator(), graph.WithLabel("~lores"))
	env.Graph().AddEdge(gen.NodeID, nodeID, "in")
	env.Graph().AddEdge(cutoff.NodeID, nodeID, "cutoff")
	env.Graph().AddEdge(resonance.NodeID, nodeID, "resonance")

	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func asGen(env value.Environment, v value.Value) (*value.Gen, bool) {
	// asGen converts a Value to a Gen, if possible.
	switch v := v.(type) {
	case *value.Gen:
		return v, true
	case *value.Num:
		id := env.Graph().AddGeneratorNode(generator.NewConstant(v.Value), graph.WithLabel(fmt.Sprintf("%v", v.Value)))
		return &value.Gen{
			NodeID: id,
		}, true
	default:
		return nil, false
	}
}

func asList(v value.Value) ([]value.Value, bool) {
	// asList converts a Value to a list, if possible.
	switch v := v.(type) {
	case *value.List:
		return v.Items, true
	default:
		return nil, false
	}
}
