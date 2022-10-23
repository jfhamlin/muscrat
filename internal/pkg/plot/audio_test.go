package plot

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDFTHistogramString(t *testing.T) {
	type testCase struct {
		name          string
		data          []complex128
		sampleRate    float64
		width, height int
		opts          []PlotOption
		want          string
	}

	testCases := []testCase{}

	// read all *.in and *.out files in testdata/ascii_spectrogram
	// and create a testCase for each pair.
	//
	// The *.in file is structured as follows:
	// - The first line contains the sample rate as a float and the
	//   chart width and height as two integers.
	// - Subsequent lines contain the real part of a complex number.
	// The *.out file contains the expected output of the chart.

	paths, err := filepath.Glob("testdata/ascii_spectrogram/*.in")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".in")
		in, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		out, err := os.ReadFile(filepath.Join("testdata/ascii_spectrogram", name+".out"))
		if err != nil {
			t.Fatal(err)
		}

		lines := strings.Split(string(in), "\n")
		fields := strings.Fields(lines[0])
		// parse sample rate
		sampleRate, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			t.Fatal(err)
		}

		// parse width and height
		width, err := strconv.Atoi(fields[1])
		if err != nil {
			t.Fatal(err)
		}
		height, err := strconv.Atoi(fields[2])
		if err != nil {
			t.Fatal(err)
		}

		var opts []PlotOption
		for _, field := range fields[3:] {
			switch field {
			case "logDomain":
				opts = append(opts, WithLogDomain())
			case "logRange":
				opts = append(opts, WithLogRange())
			default:
				t.Fatalf("unknown option %q", field)
			}
		}

		// parse data
		data := []complex128{}
		for i, line := range lines[1:] {
			if strings.TrimSpace(line) == "" {
				continue
			}
			f, err := strconv.ParseFloat(line, 64)
			if err != nil {
				t.Fatalf("error parsing float '%s' on line %d: %v", line, i+2, err)
			}
			data = append(data, complex(f, 0))
		}
		testCases = append(testCases, testCase{
			name:       name,
			data:       data,
			sampleRate: sampleRate,
			width:      width,
			height:     height,
			opts:       opts,
			want:       string(out),
		})
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := DFTHistogramString(tc.data, tc.sampleRate, tc.width, tc.height, tc.opts...)
			assert.Equal(t, tc.want, got)
		})
	}
}
