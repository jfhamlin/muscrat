package mratlang

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/jfhamlin/muscrat/internal/pkg/generator"
	"github.com/jfhamlin/muscrat/internal/pkg/generator/stochastic"
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
				// list/vector functions
				funcSymbol("list", listBuiltin),
				funcSymbol("vector", vectorBuiltin),
				funcSymbol("length", lengthBuiltin),
				funcSymbol("conj", conjBuiltin),
				funcSymbol("concat", concatBuiltin),
				funcSymbol("first", firstBuiltin),
				funcSymbol("rest", restBuiltin),
				funcSymbol("subvec", subvecBuiltin),
				// math functions
				funcSymbol("pow", powBuiltin),
				funcSymbol("floor", floorBuiltin),
				funcSymbol("*", mulBuiltin),
				funcSymbol("/", divBuiltin),
				funcSymbol("+", addBuiltin),
				funcSymbol("-", subBuiltin),
				funcSymbol("<", ltBuiltin),
				funcSymbol(">", gtBuiltin),
				// function application
				funcSymbol("apply", applyBuiltin),
				// test predicates
				funcSymbol("eq?", eqBuiltin),
				funcSymbol("list?", isListBuiltin),
				funcSymbol("vector?", isVectorBuiltin),
				funcSymbol("seq?", isSeqBuiltin),
				funcSymbol("seqable?", isSeqableBuiltin),
				funcSymbol("empty?", emptyBuiltin),
				funcSymbol("not-empty?", notEmptyBuiltin),
				// boolean functions
				funcSymbol("not", notBuiltin),
				// plumbing
				funcSymbol("*pipe", pipeBuiltin),
				funcSymbol("pipeset", pipesetBuiltin),
				// ugen
				funcSymbol("ugen", ugenBuiltin),
				// loading sample files
				funcSymbol("open-file", openFileBuiltin),
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
				funcSymbol("rand", randBuiltin),
			},
		},
		&Package{
			Name: "mrat.osc",
			Symbols: []*Symbol{
				funcSymbol("sin", sinBuiltin),
				funcSymbol("saw", sawBuiltin),
				funcSymbol("sqr", sqrBuiltin),
				funcSymbol("tri", triBuiltin),
				funcSymbol("pulse", pulseBuiltin),
				funcSymbol("phasor", phasorBuiltin),
				funcSymbol("noise", noiseBuiltin),
				funcSymbol("pink-noise", pinkNoiseBuiltin),
				funcSymbol("sampler", samplerBuiltin),
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
				funcSymbol("*lores", loresBuiltin),
			},
		},
	}
}

func addBuiltins(env *environment) {
	for _, pkg := range builtinPackages {
		for _, sym := range pkg.Symbols {
			name := pkg.Name + "/" + sym.Name
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

func vectorBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	return value.NewVector(args), nil
}

func lengthBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("length expects 1 argument, got %v", len(args))
	}

	if args[0] == nil {
		return value.NewNum(0), nil
	}

	if c, ok := args[0].(value.Counter); ok {
		return value.NewNum(float64(c.Count())), nil
	}

	enum, ok := args[0].(value.Enumerable)
	if !ok {
		return nil, fmt.Errorf("length expects an enumerable, got %v", args[0])
	}

	ch, cancel := enum.Enumerate()
	defer cancel()

	var count int
	for range ch {
		count++
	}
	return value.NewNum(float64(count)), nil
}

func conjBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("conj expects at least 2 arguments, got %v", len(args))
	}

	conjer, ok := args[0].(value.Conjer)
	if !ok {
		return nil, fmt.Errorf("conj expects a conjer, got %v", args[0])
	}

	return conjer.Conj(args[1:]...), nil
}

func concatBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	enums := make([]value.Enumerable, len(args))
	for i, arg := range args {
		e, ok := arg.(value.Enumerable)
		if !ok {
			return nil, fmt.Errorf("concat arg %d is not enumerable: %v", i, arg)
		}
		enums[i] = e
	}

	enumerable := func() (<-chan value.Value, func()) {
		ch := make(chan value.Value)
		done := make(chan struct{})
		cancel := func() {
			close(done)
		}

		go func() {
			defer close(ch)
			for _, enum := range enums {
				select {
				case <-done:
					return
				default:
				}

				func() { // scope for defer
					eCh, eCancel := enum.Enumerate()
					defer eCancel()
					for v := range eCh {
						select {
						case ch <- v:
						case <-done:
							return
						}
					}
				}()
			}
		}()

		return ch, cancel
	}

	return &value.Seq{
		Enumerable: value.EnumerableFunc(enumerable),
	}, nil
}

func firstBuiltin(env value.Environment, args []value.Value) (out value.Value, err error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("first expects 1 argument, got %v", len(args))
	}

	if args[0] == nil {
		return nil, nil
	}

	switch c := args[0].(type) {
	case *value.List:
		if c.IsEmpty() {
			return value.NilValue, nil
		}
		return c.Item(), nil
	case *value.Vector:
		if c.Count() == 0 {
			return value.NilValue, nil
		}
		return c.ValueAt(0), nil
	}

	enum, ok := args[0].(value.Enumerable)
	if !ok {
		return nil, fmt.Errorf("first expects an enumerable, got %v", args[0])
	}

	itemCh, cancel := enum.Enumerate()
	defer cancel()

	return <-itemCh, nil
}

func restBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rest expects 1 argument, got %v", len(args))
	}

	switch c := args[0].(type) {
	case *value.List:
		if c.IsEmpty() {
			return c, nil
		}
		return c.Next(), nil
	case *value.Vector:
		if c.Count() == 0 {
			return c, nil
		}
		return c.SubVector(1, c.Count()), nil
	}

	enum, ok := args[0].(value.Enumerable)
	if !ok {
		return nil, fmt.Errorf("rest expects an enumerable, got %v", args[0])
	}

	items := []value.Value{}
	itemCh, cancel := enum.Enumerate()
	defer cancel()

	// skip the first item
	<-itemCh
	for item := range itemCh {
		items = append(items, item)
	}

	// TODO: here and elsewhere, use a Sequence/Seq value type to
	// represent a lazy sequence of values, and use that instead of a
	// List/Vector.
	return value.NewList(items), nil
}

func subvecBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) < 2 || len(args) > 3 {
		return nil, fmt.Errorf("subvec expects 2 or 3 arguments, got %v", len(args))
	}

	v, ok := args[0].(*value.Vector)
	if !ok {
		return nil, fmt.Errorf("subvec expects a vector as its first argument, got %v", args[0])
	}

	start, ok := args[1].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("subvec expects a number as its second argument, got %v", args[1])
	}

	startIdx := int(start.Value)
	endIdx := v.Count()

	if len(args) == 3 {
		end, ok := args[2].(*value.Num)
		if !ok {
			return nil, fmt.Errorf("subvec expects a number as its third argument, got %v", args[2])
		}
		endIdx = int(end.Value)
	}

	if startIdx < 0 || startIdx > v.Count() || endIdx < 0 || endIdx > v.Count() {
		return nil, fmt.Errorf("subvec indices out of bounds: %v %v", startIdx, endIdx)
	}

	return v.SubVector(startIdx, endIdx), nil
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

func isListBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("list? expects 1 argument, got %v", len(args))
	}
	_, ok := args[0].(*value.List)
	return value.NewBool(ok), nil
}

func isVectorBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("vector? expects 1 argument, got %v", len(args))
	}
	_, ok := args[0].(*value.Vector)
	return value.NewBool(ok), nil
}

func isSeqBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("seq? expects 1 argument, got %v", len(args))
	}
	_, ok := args[0].(*value.Seq)
	return value.NewBool(ok), nil
}

func isSeqableBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("seqable? expects 1 argument, got %v", len(args))
	}
	_, ok := args[0].(value.Enumerable)
	return value.NewBool(ok), nil
}

func emptyBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("empty? expects 1 argument, got %v", len(args))
	}

	switch c := args[0].(type) {
	case *value.List:
		return value.NewBool(c.IsEmpty()), nil
	case *value.Vector:
		return value.NewBool(c.Count() == 0), nil
	}

	if c, ok := args[0].(value.Counter); ok {
		return value.NewBool(c.Count() == 0), nil
	}

	e, ok := args[0].(value.Enumerable)
	if !ok {
		return nil, fmt.Errorf("empty? expects an enumerable, got %v", args[0])
	}
	ch, cancel := e.Enumerate()
	defer cancel()
	// TODO: take a context.Context to support cancelation/timeout.
	_, ok = <-ch
	return value.NewBool(!ok), nil
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

func floorBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("floor expects 1 argument, got %v", len(args))
	}
	a, ok := args[0].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("floor expects a number, got %v", args[0])
	}
	return value.NewNum(math.Floor(a.Value)), nil
}

func mulBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	var coeff float64 = 1
	gens := []*value.Gen{}
	vecs := []*value.Vector{}

	// multiply all number arguments together. the operation can take
	// vectors or generators as arguments, but not both.
	for _, arg := range args {
		switch arg := arg.(type) {
		case *value.Num:
			coeff *= arg.Value
		case *value.Gen:
			gens = append(gens, arg)
			if len(vecs) > 0 {
				return nil, fmt.Errorf("cannot multiply generators and vectors")
			}
		case *value.Vector:
			vecs = append(vecs, arg)
			if len(gens) > 0 {
				return nil, fmt.Errorf("cannot multiply generators and vectors")
			}
		default:
			return nil, fmt.Errorf("invalid type for *: %v", arg)
		}
	}

	switch {
	case len(vecs) > 0:
		res := make([]value.Value, vecs[0].Count())
		for i := range res {
			res[i] = value.NewNum(coeff)
		}
		for _, vec := range vecs {
			if vec.Count() != len(res) {
				return nil, fmt.Errorf("cannot multiply vectors of different lengths (%v and %v)", len(res), vec.Count())
			}
			for i := 0; i < vec.Count(); i++ {
				n, ok := vec.ValueAt(i).(*value.Num)
				if !ok {
					return nil, fmt.Errorf("cannot multiply vectors of non-numbers")
				}
				res[i].(*value.Num).Value *= n.Value
			}
		}
		return value.NewVector(res), nil
	case len(gens) == 0:
		return value.NewNum(coeff), nil
	}

	// Otherwise, we have a generator and (possibly) a coefficent.

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

func gtBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("> expects 2 arguments, got %v", len(args))
	}
	a, ok := args[0].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("> expects a number as the first argument, got %v", args[0])
	}
	b, ok := args[1].(*value.Num)
	if !ok {
		return nil, fmt.Errorf("> expects a number as the second argument, got %v", args[1])
	}

	return value.NewBool(a.Value > b.Value), nil
}

func ugenBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("ugen expects 1 argument, got %v", len(args))
	}
	gen, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("ugen expects a generator as the first argument, got %v", args[0])
	}
	return gen, nil
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
	nodeID := env.Graph().AddGeneratorNode(&pipe{}, graph.WithLabel("*pipe"))
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
	// the first argument should be an applyer, the second an enumerable
	applyer, ok := args[0].(value.Applyer)
	if !ok {
		return nil, fmt.Errorf("apply expects a function as the first argument, got %v", args[0])
	}
	enum, ok := args[1].(value.Enumerable)
	if !ok {
		return nil, fmt.Errorf("apply expects an enumerable as the second argument, got %v", args[1])
	}
	values, err := value.EnumerateAll(context.Background(), enum)
	if err != nil {
		return nil, err
	}
	return applyer.Apply(env, values)
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

func randBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("rand expects 0 arguments, got %v", len(args))
	}

	return &value.Num{Value: rand.Float64()}, nil
}

func phasorBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
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
			return nil, fmt.Errorf("invalid type for phasor frequency: %v", arg)
		}
	}
	nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Phasor(1024)), graph.WithLabel("phasor"))
	env.Graph().AddEdge(freq, nodeID, "w")
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func handleExtraGenArgs(env value.Environment, nodeID graph.NodeID, args []value.Value) error {
	for len(args) > 0 {
		if len(args) < 2 {
			return fmt.Errorf("expected key-value pairs in extra arguments, got %v", args)
		}
		key, val := args[0], args[1]
		args = args[2:]

		kw, ok := key.(*value.Keyword)
		if !ok {
			return fmt.Errorf("expected keyword as key, got %v", key)
		}
		switch kw.Value {
		case "iphase":
			_, ok := val.(*value.Num)
			if !ok {
				return fmt.Errorf("expected number as iphase value, got %v", val)
			}
			gen, _ := asGen(env, val)
			env.Graph().AddEdge(gen.NodeID, nodeID, "iphase")
		case "phase":
			phase, ok := asGen(env, val)
			if !ok {
				return fmt.Errorf("expected generator as phase, got %v", val)
			}
			env.Graph().AddEdge(phase.NodeID, nodeID, "phase")
		case "sync":
			sync, ok := asGen(env, val)
			if !ok {
				return fmt.Errorf("expected generator as sync, got %v", val)
			}
			env.Graph().AddEdge(sync.NodeID, nodeID, "sync")
		case "duty":
			duty, ok := asGen(env, val)
			if !ok {
				return fmt.Errorf("expected generator as duty cycle, got %v", val)
			}
			env.Graph().AddEdge(duty.NodeID, nodeID, "dc")
		default:
			return fmt.Errorf("unknown key %v", kw.Value)
		}
	}
	return nil
}

func sinBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Sin(1024)), graph.WithLabel("sin"))
	if len(args) == 0 {
		return &value.Gen{
			NodeID: nodeID,
		}, nil
	}

	freq, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("expected generator or number as the first argument, got %v", args[0])
	}
	env.Graph().AddEdge(freq.NodeID, nodeID, "w")
	args = args[1:]

	if err := handleExtraGenArgs(env, nodeID, args); err != nil {
		return nil, err
	}

	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func sawBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Saw(1024)), graph.WithLabel("saw"))
	if len(args) == 0 {
		return &value.Gen{
			NodeID: nodeID,
		}, nil
	}

	freq, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("expected generator or number as the first argument, got %v", args[0])
	}
	env.Graph().AddEdge(freq.NodeID, nodeID, "w")
	args = args[1:]

	if err := handleExtraGenArgs(env, nodeID, args); err != nil {
		return nil, err
	}

	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func triBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Tri(1024)), graph.WithLabel("tri"))
	if len(args) == 0 {
		return &value.Gen{
			NodeID: nodeID,
		}, nil
	}

	freq, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("expected generator or number as the first argument, got %v", args[0])
	}
	env.Graph().AddEdge(freq.NodeID, nodeID, "w")
	args = args[1:]

	if err := handleExtraGenArgs(env, nodeID, args); err != nil {
		return nil, err
	}

	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func sqrBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Pulse(1024), wavtabs.WithDefaultDutyCycle(0.5)), graph.WithLabel("sqr"))
	if len(args) == 0 {
		return &value.Gen{
			NodeID: nodeID,
		}, nil
	}

	freq, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("invalid type for frequency: %v", args[0])
	}
	env.Graph().AddEdge(freq.NodeID, nodeID, "w")
	args = args[1:]
	if len(args) == 0 {
		return &value.Gen{
			NodeID: nodeID,
		}, nil
	}

	dutyCycle, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("invalid type for duty cycle: %v", args[0])
	}
	env.Graph().AddEdge(dutyCycle.NodeID, nodeID, "dc")
	args = args[1:]

	if err := handleExtraGenArgs(env, nodeID, args); err != nil {
		return nil, err
	}

	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func pulseBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	nodeID := env.Graph().AddGeneratorNode(wavtabs.Generator(wavtabs.Pulse(1024), wavtabs.WithDefaultDutyCycle(0.5)), graph.WithLabel("pulse"))
	if len(args) == 0 {
		return &value.Gen{
			NodeID: nodeID,
		}, nil
	}

	freq, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("invalid type for frequency: %v", args[0])
	}
	env.Graph().AddEdge(freq.NodeID, nodeID, "w")
	args = args[1:]
	if len(args) == 0 {
		return &value.Gen{
			NodeID: nodeID,
		}, nil
	}

	dutyCycle, ok := asGen(env, args[0])
	if !ok {
		return nil, fmt.Errorf("invalid type for duty cycle: %v", args[0])
	}
	env.Graph().AddEdge(dutyCycle.NodeID, nodeID, "dc")
	args = args[1:]

	if err := handleExtraGenArgs(env, nodeID, args); err != nil {
		return nil, err
	}

	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func outBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	// add a new sink node for every argument. each sink represents an output channel.
	numSinks := len(env.Graph().Sinks())
	// if len(sinks) == 0 {
	// 	sinks = append(sinks, env.Graph().AddSinkNode(graph.WithLabel("out")))
	// }
	for i, arg := range args {
		gen, ok := asGen(env, arg)
		if !ok {
			return nil, fmt.Errorf("expected generator, got %v", arg)
		}

		chanID := i + numSinks
		sink := env.Graph().AddSinkNode(graph.WithLabel(fmt.Sprintf("out%d", chanID)))
		env.Graph().AddEdge(gen.NodeID, sink.ID(), "w")
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

func NewClipGenerator() generator.SampleGenerator {
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		input := cfg.InputSamples["$0"]
		mins := cfg.InputSamples["min"]
		maxs := cfg.InputSamples["max"]
		output := make([]float64, n)
		for i := 0; i < n; i++ {
			min := mins[0]
			max := maxs[0]
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
	min, ok := asGen(env, args[1])
	if !ok {
		return nil, fmt.Errorf("clip expects a gennable as the second argument, got %v", args[1])
	}
	max, ok := asGen(env, args[2])
	if !ok {
		return nil, fmt.Errorf("clip expects a gennable as the third argument, got %v", args[2])
	}
	nodeID := env.Graph().AddGeneratorNode(NewClipGenerator(), graph.WithLabel("clip"))
	env.Graph().AddEdge(gen.NodeID, nodeID, "$0")
	env.Graph().AddEdge(min.NodeID, nodeID, "min")
	env.Graph().AddEdge(max.NodeID, nodeID, "max")
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
		}

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

func pinkNoiseBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("pinkNoise expects 0 arguments, got %v", len(args))
	}
	nodeID := env.Graph().AddGeneratorNode(stochastic.NewPinkNoise(), graph.WithLabel("pink-noise"))
	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func NewSamplerGenerator(v *value.Vector, loop bool) generator.SampleGenerator {
	if v.Count() == 0 {
		return generator.NewConstant(0)
	}

	sampleLen := v.Count()
	index := 0
	stopped := false
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		res := make([]float64, n)
		for i := 0; i < n; i++ {
			if stopped {
				res[i] = 0
				continue
			}

			// TODO: do this once and cache it.
			if num, ok := v.ValueAt(index).(*value.Num); ok {
				res[i] = num.Value
			} else {
				res[i] = 0
			}
			index++
			if index >= sampleLen {
				index = 0
				if !loop {
					stopped = true
				}
			}
		}
		return res
	})
}

func gatherArgsAndFlags(args []value.Value, allowedKeywords ...string) ([]value.Value, map[string]value.Value, error) {
	posArgs := make([]value.Value, 0, len(args))
	flags := make(map[string]value.Value)
	for len(args) > 0 {
		if _, ok := args[0].(*value.Keyword); ok {
			// handle keyword arguments
			break
		}
		posArgs = append(posArgs, args[0])
		args = args[1:]
	}
	if len(args)%2 != 0 {
		return nil, nil, fmt.Errorf("expected even number of keyword arguments, got %v", len(args))
	}
	for i := 0; i < len(args); i += 2 {
		kw, ok := args[i].(*value.Keyword)
		if !ok {
			return nil, nil, fmt.Errorf("expected keyword argument, got %v", args[i])
		}
		if !containsString(allowedKeywords, kw.Value) {
			return nil, nil, fmt.Errorf("unknown keyword argument %v", kw.Value)
		}
		flags[kw.Value] = args[i+1]
	}
	return posArgs, flags, nil
}

func samplerBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	args, flags, err := gatherArgsAndFlags(args, "loop")
	if err != nil {
		return nil, err
	}

	// sampler takes a required vector argument
	// TODO: sample rate, gate
	if len(args) < 1 {
		return nil, fmt.Errorf("sampler expects at least 1 argument, got %v", len(args))
	}
	vec, ok := args[0].(*value.Vector)
	if !ok {
		return nil, fmt.Errorf("sampler expects a vector as the first argument, got %v", args[0])
	}

	loop := false
	if loopFlag, ok := flags["loop"]; ok {
		loop = value.IsTruthy(loopFlag)
	}

	nodeID := env.Graph().AddGeneratorNode(NewSamplerGenerator(vec, loop), graph.WithLabel("sampler"))
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
	levels, ok := asSlice(args[1])
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
	durations, ok := asSlice(args[2])
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
		return nil, fmt.Errorf("*lores expects 3 arguments, got %v", len(args))
	}

	gen, ok := args[0].(*value.Gen)
	if !ok {
		return nil, fmt.Errorf("*lores expects a Gen as the first argument, got %v", args[0])
	}

	cutoff, ok := asGen(env, args[1])
	if !ok {
		return nil, fmt.Errorf("*lores expects a gennable as the second argument, got %v", args[1])
	}
	resonance, ok := asGen(env, args[2])
	if !ok {
		return nil, fmt.Errorf("*lores expects a gennable as the third argument, got %v", args[2])
	}

	nodeID := env.Graph().AddGeneratorNode(NewLowpassFilterGenerator(), graph.WithLabel("*lores"))
	env.Graph().AddEdge(gen.NodeID, nodeID, "in")
	env.Graph().AddEdge(cutoff.NodeID, nodeID, "cutoff")
	env.Graph().AddEdge(resonance.NodeID, nodeID, "resonance")

	return &value.Gen{
		NodeID: nodeID,
	}, nil
}

func openFileBuiltin(env value.Environment, args []value.Value) (value.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("open-file expects 1 argument, got %v", len(args))
	}

	filename, ok := args[0].(*value.Str)
	if !ok {
		return nil, fmt.Errorf("open-file expects a string as the first argument, got %v", args[0])
	}

	f, err := os.Open(filename.Value)
	if err != nil {
		return nil, fmt.Errorf("open-file: error opening file: %v", err)
	}
	defer f.Close()

	dec := wav.NewDecoder(f)
	if !dec.IsValidFile() {
		return nil, fmt.Errorf("open-file: file '%s' is not a valid WAV file", filename.Value)
	}

	var intSamples []int
	audioBuf := &audio.IntBuffer{Data: make([]int, 2048)}
	for {
		n, err := dec.PCMBuffer(audioBuf)
		if err != nil {
			return nil, fmt.Errorf("open-file: error reading PCM data: %v", err)
		}
		if n == 0 {
			break
		}
		intSamples = append(intSamples, audioBuf.Data...)
	}
	bitDepth := dec.SampleBitDepth()

	floatSamples := make([]float64, len(intSamples))

	for _, s := range intSamples {
		floatSample := float64(s) / float64(int(1)<<uint(bitDepth-1))
		if floatSample > 1 {
			floatSample = 1
		} else if floatSample < -1 {
			floatSample = -1
		}
		floatSamples = append(floatSamples, floatSample)
	}

	// resample to 44100 Hz, assumed to be the sample rate of the audio device
	// TODOs:
	// - make this configurable
	// - don't assume 44100 Hz
	const deviceSampleRate = 44100
	if dec.SampleRate != deviceSampleRate {
		outputSamples := make([]float64, len(floatSamples)*deviceSampleRate/int(dec.SampleRate))
		for i := range outputSamples {
			t := float64(i) / float64(len(outputSamples)-1)
			outputSamples[i] = floatSamples[int(t*float64(len(floatSamples)-1))]
		}
		floatSamples = outputSamples
	}

	sampleValues := make([]value.Value, len(floatSamples))
	for i, s := range floatSamples {
		sampleValues[i] = value.NewNum(s)
	}

	return value.NewVector(sampleValues), nil
}

func enumerableToGen(env value.Environment, enum value.Enumerable) *value.Gen {
	ch, cancel := enum.Enumerate()
	isDone := false
	gf := generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		out := make([]float64, n)
		for i := 0; i < n; i++ {
			if isDone {
				out[i] = 0
				continue
			}

			if val, ok := <-ch; ok {
				num, ok := val.(*value.Num)
				if !ok {
					out[i] = 0
				} else {
					out[i] = num.Value
				}
				continue
			}
			cancel()

			isDone = true
			out[i] = 0
		}
		return out
	})

	return &value.Gen{
		NodeID: env.Graph().AddGeneratorNode(gf, graph.WithLabel(fmt.Sprintf("buffer"))),
	}
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
	case value.Enumerable:
		return enumerableToGen(env, v), true
	default:
		return nil, false
	}
}

// asSlice converts a Value to a slice of Values, if possible.
func asSlice(v value.Value) ([]value.Value, bool) {
	enum, ok := v.(value.Enumerable)
	if !ok {
		return nil, false
	}

	ch, cancel := enum.Enumerate()
	defer cancel()

	var list []value.Value
	for v := range ch {
		list = append(list, v)
	}
	return list, true
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
