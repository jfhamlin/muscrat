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
				funcSymbol("*", mulBuiltin),
				funcSymbol("/", divBuiltin),
				funcSymbol("+", addBuiltin),
				// function application
				funcSymbol("apply", applyBuiltin),
				// test predicates
				funcSymbol("eq?", eqBuiltin),
				funcSymbol("empty?", emptyBuiltin),
				funcSymbol("not-empty?", notEmptyBuiltin),
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
			Name: "mrat.osc",
			Symbols: []*Symbol{
				funcSymbol("sin", sinBuiltin),
				funcSymbol("saw", sawBuiltin),
				funcSymbol("sqr", sqrBuiltin),
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
			env.define(name, sym.Value)
		}
	}
}

func funcSymbol(name string, fn func(*environment, []Value) (Value, error)) *Symbol {
	return &Symbol{
		Name: name,
		Value: &BuiltinFunc{
			applyer: applyerFunc(fn),
			name:    name,
		},
	}
}

func loadFile(env *environment, filename string) error {
	absFile, ok := env.resolveFile(filename)
	if !ok {
		return fmt.Errorf("could not resolve file %v", filename)
	}

	buf, err := ioutil.ReadFile(absFile)
	if err != nil {
		return fmt.Errorf("error reading file %v: %w", filename, err)
	}

	prog, err := Parse(strings.NewReader(string(buf)))
	if err != nil {
		return fmt.Errorf("error parsing file %v: %w", filename, err)
	}

	loadEnv := env.pushLoadPaths([]string{filepath.Dir(absFile)})
	_, _, err = prog.Eval(withEnv(loadEnv))
	if err != nil {
		return fmt.Errorf("error evaluating file %v: %w", filename, err)
	}

	return nil
}

func loadFileBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("load-file expects 1 argument, got %v", len(args))
	}
	filename, ok := args[0].(*Str)
	if !ok {
		return nil, fmt.Errorf("load-file expects a string, got %v", args[0])
	}
	return nil, loadFile(env, filename.Value)
}

func listBuiltin(env *environment, args []Value) (Value, error) {
	return &List{Values: args}, nil
}

func lengthBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("length expects 1 argument, got %v", len(args))
	}
	switch arg := args[0].(type) {
	case *List:
		return &Num{Value: float64(len(arg.Values))}, nil
	default:
		return nil, fmt.Errorf("length expects a list, got %v", arg)
	}
}

func concatBuiltin(env *environment, args []Value) (Value, error) {
	var res []Value
	for _, arg := range args {
		switch arg := arg.(type) {
		case *List:
			res = append(res, arg.Values...)
		default:
			return nil, fmt.Errorf("invalid type for concat: %v", arg)
		}
	}
	return &List{Values: res}, nil
}

func firstBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("first expects 1 argument, got %v", len(args))
	}
	list, ok := args[0].(*List)
	if !ok {
		return nil, fmt.Errorf("first expects a list, got %v", args[0])
	}
	if len(list.Values) == 0 {
		return nil, fmt.Errorf("first expects a non-empty list, got %v", list)
	}
	return list.Values[0], nil
}

func restBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("rest expects 1 argument, got %v", len(args))
	}
	list, ok := args[0].(*List)
	if !ok {
		return nil, fmt.Errorf("rest expects a list, got %v", args[0])
	}
	if len(list.Values) == 0 {
		return list, nil
	}
	return &List{Values: list.Values[1:]}, nil
}

func notBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("not expects 1 argument, got %v", len(args))
	}
	switch arg := args[0].(type) {
	case *Bool:
		return &Bool{Value: !arg.Value}, nil
	default:
		return nil, fmt.Errorf("not expects a boolean, got %v", arg)
	}
}

func eqBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("eq? expects 2 arguments, got %v", len(args))
	}
	return &Bool{Value: args[0].Equal(args[1])}, nil
}

func emptyBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("empty? expects 1 argument, got %v", len(args))
	}
	list, ok := args[0].(*List)
	if !ok {
		return nil, fmt.Errorf("empty? expects a list, got %v", args[0])
	}
	return &Bool{Value: len(list.Values) == 0}, nil
}

func notEmptyBuiltin(env *environment, args []Value) (Value, error) {
	v, err := emptyBuiltin(env, args)
	if err != nil {
		return nil, err
	}
	return notBuiltin(env, []Value{v})
}

func mulBuiltin(env *environment, args []Value) (Value, error) {
	var coeff float64 = 1
	gens := []*Gen{}

	// multiply all number arguments together
	for _, arg := range args {
		switch arg := arg.(type) {
		case *Num:
			coeff *= arg.Value
		case *Gen:
			gens = append(gens, arg)
		default:
			return nil, fmt.Errorf("invalid type for *: %v", arg)
		}
	}
	// if there are no generators, return the result of multiplying all the numbers
	if len(gens) == 0 {
		return &Num{Value: coeff}, nil
	}

	// otherwise, create a new constant generator node for the coefficient
	constNodeID := env.graph.AddGeneratorNode(generator.NewConstant(coeff), graph.WithLabel(fmt.Sprintf("%v", coeff)))
	gens = append(gens, &Gen{NodeID: constNodeID})

	// create a new generator node for the product
	nodeID := env.graph.AddGeneratorNode(generator.NewProduct(), graph.WithLabel("*"))
	for i, gen := range gens {
		env.graph.AddEdge(gen.NodeID, nodeID, fmt.Sprintf("$%d", i))
	}
	return &Gen{
		NodeID: nodeID,
	}, nil
}

func divBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("div expects 2 arguments, got %v", len(args))
	}
	num, ok := args[0].(*Num)
	if !ok {
		return nil, fmt.Errorf("div expects a number as the first argument, got %v", args[0])
	}
	denom, ok := args[1].(*Num)
	if !ok {
		return nil, fmt.Errorf("div expects a number as the second argument, got %v", args[1])
	}
	// TODO: handle generators
	return &Num{Value: num.Value / denom.Value}, nil
}

func addBuiltin(env *environment, args []Value) (Value, error) {
	var sum float64 = 0
	gens := []*Gen{}

	// sum all number arguments together
	for _, arg := range args {
		switch arg := arg.(type) {
		case *Num:
			sum += arg.Value
		case *Gen:
			gens = append(gens, arg)
		default:
			return nil, fmt.Errorf("invalid type for +: %v", arg)
		}
	}
	// if there are no generators, return the result of summing all the numbers
	if len(gens) == 0 {
		return &Num{Value: sum}, nil
	}

	// otherwise, if the sum is not zero, create a new constant
	// generator node for the sum
	if sum != 0 {
		constNodeID := env.graph.AddGeneratorNode(generator.NewConstant(sum), graph.WithLabel(fmt.Sprintf("%v", sum)))
		gens = append(gens, &Gen{NodeID: constNodeID})
	}

	// create a new generator node for the sum
	nodeID := env.graph.AddGeneratorNode(generator.NewSum(), graph.WithLabel("+"))
	for i, gen := range gens {
		env.graph.AddEdge(gen.NodeID, nodeID, fmt.Sprintf("$%d", i))
	}
	return &Gen{
		NodeID: nodeID,
	}, nil
}

func applyBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("apply expects 2 arguments, got %v", len(args))
	}
	// the first argument should be an applyer, the second a list
	applyer, ok := args[0].(applyer)
	if !ok {
		return nil, fmt.Errorf("apply expects a function as the first argument, got %v", args[0])
	}
	list, ok := args[1].(*List)
	if !ok {
		return nil, fmt.Errorf("apply expects a list as the second argument, got %v", args[1])
	}
	return applyer.Apply(env, list.Values)
}

func printBuiltin(env *environment, args []Value) (Value, error) {
	for i, arg := range args {
		if arg == nil {
			// TODO: add nil to the type system
			env.stdout.Write([]byte("nil"))
		} else {
			env.stdout.Write([]byte(arg.String()))
		}
		if i < len(args)-1 {
			env.stdout.Write([]byte(" "))
		}
	}
	env.stdout.Write([]byte("\n"))
	return nil, nil
}

func sinBuiltin(env *environment, args []Value) (Value, error) {
	var freq graph.NodeID
	if len(args) == 0 {
		freq = env.graph.AddGeneratorNode(generator.NewConstant(440), graph.WithLabel("440"))
	} else {
		switch arg := args[0].(type) {
		case *Num:
			freq = env.graph.AddGeneratorNode(generator.NewConstant(arg.Value), graph.WithLabel(fmt.Sprintf("%v", arg.Value)))
		case *Gen:
			freq = arg.NodeID
		default:
			return nil, fmt.Errorf("invalid type for sin frequency: %v", arg)
		}
	}
	nodeID := env.graph.AddGeneratorNode(wavtabs.Generator(wavtabs.Sin(1024)), graph.WithLabel("sin"))
	env.graph.AddEdge(freq, nodeID, "w")
	return &Gen{
		NodeID: nodeID,
	}, nil
}

func sawBuiltin(env *environment, args []Value) (Value, error) {
	var freq graph.NodeID
	if len(args) == 0 {
		freq = env.graph.AddGeneratorNode(generator.NewConstant(440), graph.WithLabel("440"))
	} else {
		switch arg := args[0].(type) {
		case *Num:
			freq = env.graph.AddGeneratorNode(generator.NewConstant(arg.Value), graph.WithLabel(fmt.Sprintf("%v", arg.Value)))
		case *Gen:
			freq = arg.NodeID
		default:
			return nil, fmt.Errorf("invalid type for sin frequency: %v", arg)
		}
	}
	nodeID := env.graph.AddGeneratorNode(wavtabs.Generator(wavtabs.Saw(1024)), graph.WithLabel("saw"))
	env.graph.AddEdge(freq, nodeID, "w")
	return &Gen{
		NodeID: nodeID,
	}, nil
}

func NewSquareGenerator() generator.SampleGenerator {
	phase := 0.0
	return generator.SampleGeneratorFunc(func(ctx context.Context, cfg generator.SampleConfig, n int) []float64 {
		dcs := cfg.InputSamples["dc"]
		ws := cfg.InputSamples["w"]
		res := make([]float64, n)

		lastDC := dcs[0]
		wavtab := wavtabs.Square(1024, lastDC)
		w := 0.0
		for i := 0; i < n; i++ {
			if dcs[i] != lastDC {
				lastDC = dcs[i]
				wavtab = wavtabs.Square(1024, lastDC)
			}
			if i < len(ws) {
				w = ws[i]
			}
			res[i] = wavtab.Lerp(phase)
			phase += w / float64(cfg.SampleRateHz)
			if phase > 1 {
				phase -= 1
			}
		}
		return res
	})
}

func sqrBuiltin(env *environment, args []Value) (Value, error) {
	var freq graph.NodeID
	if len(args) == 0 {
		freq = env.graph.AddGeneratorNode(generator.NewConstant(440), graph.WithLabel("440"))
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
		dutyCycle = env.graph.AddGeneratorNode(generator.NewConstant(0.5), graph.WithLabel("0.5"))
	} else {
		gen, ok := asGen(env, args[0])
		if !ok {
			return nil, fmt.Errorf("invalid type for saw duty cycle: %v", args[0])
		}
		dutyCycle = gen.NodeID
	}

	nodeID := env.graph.AddGeneratorNode(NewSquareGenerator(), graph.WithLabel("sqr"))
	env.graph.AddEdge(freq, nodeID, "w")
	env.graph.AddEdge(dutyCycle, nodeID, "dc")
	return &Gen{
		NodeID: nodeID,
	}, nil
}

func outBuiltin(env *environment, args []Value) (Value, error) {
	sinks := env.sinks
	if len(sinks.sinkNodeIDs) == 0 {
		id, ch := env.graph.AddSinkNode(graph.WithLabel("out"))
		sinks.sinkNodeIDs = append(sinks.sinkNodeIDs, id)
		sinks.sinkChannels = append(sinks.sinkChannels, ch)
	}
	for _, arg := range args {
		switch arg := arg.(type) {
		case *Gen:
			for _, id := range sinks.sinkNodeIDs {
				env.graph.AddEdge(arg.NodeID, id, arg.String())
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

func freeverbBuiltin(env *environment, args []Value) (Value, error) {
	// take a single argument, which should be a Gen
	if len(args) != 1 {
		return nil, fmt.Errorf("freeverb expects 1 argument, got %v", len(args))
	}
	gen, ok := args[0].(*Gen)
	if !ok {
		return nil, fmt.Errorf("freeverb expects a Gen as the first argument, got %v", args[0])
	}
	nodeID := env.graph.AddGeneratorNode(NewFreeverbGenerator(), graph.WithLabel("freeverb"))
	env.graph.AddEdge(gen.NodeID, nodeID, "$0")
	return &Gen{
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

func clipBuiltin(env *environment, args []Value) (Value, error) {
	// clip takes three arguments: a Gen, a min, and a max
	if len(args) != 3 {
		return nil, fmt.Errorf("clip expects 3 arguments, got %v", len(args))
	}
	gen, ok := args[0].(*Gen)
	if !ok {
		return nil, fmt.Errorf("clip expects a Gen as the first argument, got %v", args[0])
	}
	min, ok := args[1].(*Num)
	if !ok {
		return nil, fmt.Errorf("clip expects a Num as the second argument, got %v", args[1])
	}
	max, ok := args[2].(*Num)
	if !ok {
		return nil, fmt.Errorf("clip expects a Num as the third argument, got %v", args[2])
	}
	nodeID := env.graph.AddGeneratorNode(NewClipGenerator(min.Value, max.Value), graph.WithLabel("clip"))
	env.graph.AddEdge(gen.NodeID, nodeID, "$0")
	return &Gen{
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

func delayBuiltin(env *environment, args []Value) (Value, error) {
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
	nodeID := env.graph.AddGeneratorNode(NewDelayGenerator(), graph.WithLabel("delay"))
	env.graph.AddEdge(gen.NodeID, nodeID, "$0")
	env.graph.AddEdge(delay.NodeID, nodeID, "delay")
	return &Gen{
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

func mixerBuiltin(env *environment, args []Value) (Value, error) {
	// mixer takes a variable number of arguments that can be converted to generators
	// and mixes them together.
	gens := make([]*Gen, len(args))
	for i, arg := range args {
		gen, ok := asGen(env, arg)
		if !ok {
			return nil, fmt.Errorf("mixer expects a Gen as the %v argument, got %v", i, arg)
		}
		gens[i] = gen
	}
	nodeID := env.graph.AddGeneratorNode(NewMixerGenerator(), graph.WithLabel("mixer"))
	for i, gen := range gens {
		env.graph.AddEdge(gen.NodeID, nodeID, fmt.Sprintf("$%v", i))
	}
	return &Gen{
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

func noiseBuiltin(env *environment, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, fmt.Errorf("noise expects 0 arguments, got %v", len(args))
	}
	nodeID := env.graph.AddGeneratorNode(Noise, graph.WithLabel("noise"))
	return &Gen{
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

func NewEnvelopeGenerator() generator.SampleGenerator {
	// Behavior follows that of SuperCollider's Env/EnvGen
	// https://doc.sccode.org/Classes/Env.html

	fmt.Println("new envelope generator")

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
					// interpolate between levels[j] and levels[j+1]
					// at time triggerTime
					res[i] = levels[j][i] + (levels[j+1][i]-levels[j][i])*(triggerTime-(timeSum-t[i]))/t[i]
					break
				}
			}
			triggerTime += 1 / float64(cfg.SampleRateHz)
		}
		return res
	})
}

func envBuiltin(env *environment, args []Value) (Value, error) {
	// env takes three arguments:
	//
	// 1. a gating signal. when the signal is > 0, the envelope is triggered.
	//
	// 2. a list of levels, which are the values of the envelope at each
	// point. these must be convertible to Gens.
	//
	// 3. a list of durations, which are the durations of each segment
	// of the envelope. these must be convertible to Gens.

	if len(args) != 3 {
		return nil, fmt.Errorf("env expects 3 arguments, got %v", len(args))
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
	levelGens := make([]*Gen, len(levels))
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
	durationGens := make([]*Gen, len(durations))
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

	// create the envelope generator.
	nodeID := env.graph.AddGeneratorNode(NewEnvelopeGenerator(), graph.WithLabel("env"))
	// nodeID := env.graph.AddGeneratorNode(generator.NewConstant(1), graph.WithLabel("env"))
	env.graph.AddEdge(trigger.NodeID, nodeID, "trigger")
	for i, gen := range levelGens {
		env.graph.AddEdge(gen.NodeID, nodeID, fmt.Sprintf("level$%v", i))
	}
	for i, gen := range durationGens {
		env.graph.AddEdge(gen.NodeID, nodeID, fmt.Sprintf("time$%v", i))
	}

	return &Gen{
		NodeID: nodeID,
	}, nil
}

func asGen(env *environment, v Value) (*Gen, bool) {
	// asGen converts a Value to a Gen, if possible.
	switch v := v.(type) {
	case *Gen:
		return v, true
	case *Num:
		id := env.graph.AddGeneratorNode(generator.NewConstant(v.Value), graph.WithLabel(fmt.Sprintf("%v", v.Value)))
		return &Gen{
			NodeID: id,
		}, true
	default:
		return nil, false
	}
}

func asList(v Value) ([]Value, bool) {
	// asList converts a Value to a list, if possible.
	switch v := v.(type) {
	case *List:
		return v.Values, true
	default:
		return nil, false
	}
}
