package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/jfhamlin/muscrat/internal/pkg/plot"
	"github.com/jfhamlin/muscrat/pkg/wavtabs"
)

const (
	resolution = wavtabs.DefaultResolution
)

var (
	tables = map[string]*wavtabs.Table{
		"sin":    wavtabs.Sin(resolution),
		"saw":    wavtabs.Saw(resolution),
		"tri":    wavtabs.Tri(resolution),
		"phasor": wavtabs.Phasor(resolution),
		"pulse":  wavtabs.Pulse(resolution),
	}

	output = flag.String("o", "", "output file")
)

func main() {
	flag.Parse()

	outFile := os.Stdout
	if *output != "" {
		var err error
		outFile, err = os.Create(*output)
		if err != nil {
			panic(err)
		}
		defer outFile.Close()
	}

	names := make([]string, 0, len(tables))
	for name := range tables {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintln(outFile, "=", name, strings.Repeat("=", 120))
		table := tables[name]
		fmt.Fprintln(outFile, "nearest")
		fmt.Fprintln(outFile, plot.LineChartString(tblValsNearest(table), 120, 40))
		fmt.Fprintln(outFile, "linear")
		fmt.Fprintln(outFile, plot.LineChartString(tblValsLinear(table), 120, 40))
		fmt.Fprintln(outFile, "hermite")
		fmt.Fprintln(outFile, plot.LineChartString(tblValsHermite(table), 120, 40))
	}
}

func tblValsNearest(tbl *wavtabs.Table) []float64 {
	vals := make([]float64, 0, 1000)
	for i := 0.0; i <= 2.0; i += 0.001 {
		vals = append(vals, tbl.Nearest(i))
	}
	return vals
}

func tblValsLinear(tbl *wavtabs.Table) []float64 {
	vals := make([]float64, 0, 1000)
	for i := 0.0; i <= 2.0; i += 0.001 {
		vals = append(vals, tbl.Lerp(i))
	}
	return vals
}

func tblValsHermite(tbl *wavtabs.Table) []float64 {
	vals := make([]float64, 0, 1000)
	for i := 0.0; i <= 2.0; i += 0.001 {
		vals = append(vals, tbl.Hermite(i))
	}
	return vals
}
