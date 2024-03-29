package plot

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLineChartString(t *testing.T) {
	type testCase struct {
		name          string
		data          []float64
		width, height int
		want          string
	}

	testCases := []testCase{}

	// read all *.in and *.out files in testdata/ascii_chart
	// and create a testCase for each pair.
	//
	// The *.in file is structured as follows:
	// - The first line contains the chart width and height as two integers.
	// - Subsequent lines contain the data to plot as newline-separated floats.
	// The *.out file contains the expected output of the chart.

	paths, err := filepath.Glob("testdata/ascii_chart/*.in")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".in")
		in, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		out, err := os.ReadFile(filepath.Join("testdata/ascii_chart", name+".out"))
		if err != nil {
			t.Fatal(err)
		}
		// parse width and height
		lines := strings.Split(string(in), "\n")
		widthHeightFields := strings.Fields(lines[0])
		width, err := strconv.Atoi(widthHeightFields[0])
		if err != nil {
			t.Fatal(err)
		}
		height, err := strconv.Atoi(widthHeightFields[1])
		if err != nil {
			t.Fatal(err)
		}

		// parse data
		data := []float64{}
		for i, line := range lines[1:] {
			if strings.TrimSpace(line) == "" {
				continue
			}
			f, err := strconv.ParseFloat(line, 64)
			if err != nil {
				t.Fatalf("error parsing float '%s' on line %d: %v", line, i+2, err)
			}
			data = append(data, f)
		}
		testCases = append(testCases, testCase{
			name:   name,
			data:   data,
			width:  width,
			height: height,
			want:   string(out),
		})
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := LineChartString(tc.data, tc.width, tc.height)
			assert.Equal(t, tc.want, got)
		})
	}
}
