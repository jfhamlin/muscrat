package mratlang

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	type testCase struct {
		name   string
		input  string
		output string
	}

	var testCases = []testCase{}

	// read all *.in files in testdata/parser as test cases.
	paths, err := filepath.Glob("testdata/parser/*.in")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// read corresponding *.out file.
		outPath := strings.TrimSuffix(path, ".in") + ".out"
		outData, err := ioutil.ReadFile(outPath)
		if err != nil {
			t.Fatal(err)
		}
		testCases = append(testCases, testCase{
			name:   filepath.Base(path),
			input:  string(data),
			output: string(outData),
		})
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			prog, err := Parse(strings.NewReader(tc.input), WithFilename(tc.name))
			if err != nil {
				t.Fatal(err)
			}

			// save stdout to a buffer
			stdout := &strings.Builder{}

			_, _, err = prog.Eval(WithStdout(stdout), WithLoadPath([]string{"testdata/parser"}))
			if err != nil {
				t.Fatal(err)
			}

			if got, want := stdout.String(), tc.output; got != want {
				t.Errorf("\n=== got ====\n%s============\n=== want ===\n%s============", got, want)
			}
		})
	}
}
