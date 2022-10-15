package value

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jfhamlin/muscrat/internal/pkg/mratlang/reader"
)

func TestFromAST(t *testing.T) {
	type testCase struct {
		name   string
		input  string
		output string
	}

	var testCases = []testCase{}
	// read all *.mrat files in the reader's tests as test cases.
	paths, err := filepath.Glob("../reader/testdata/reader/*.mrat")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		// read corresponding *.out file.
		outPath := strings.TrimSuffix(path, ".mrat") + ".out"
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
			r := reader.New(strings.NewReader(tc.input))
			exprs, err := r.ReadAll()
			if err != nil {
				t.Fatal(err)
			}

			strs := make([]string, len(exprs)+1)
			strs[len(strs)-1] = ""
			for i, expr := range exprs {
				strs[i] = FromAST(expr).String()
			}
			output := strings.Join(strs, "\n")
			if output != tc.output {
				t.Errorf("output mismatch:\nwant:\n%s\nhave:\n%s[END]\n", tc.output, output)
			}
		})
	}
}

func FuzzFromAST(f *testing.F) {
	paths, err := filepath.Glob("../reader/testdata/reader/*.mrat")
	if err != nil {
		f.Fatal(err)
	}
	for _, path := range paths {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(string(data))
	}
	f.Fuzz(func(t *testing.T, program string) {
		r := reader.New(strings.NewReader(program))
		exprs, err := r.ReadAll()
		if err != nil {
			// ignore errors. if the program is invalid, we won't bother
			// comparing outputs.
			return
		}
		for _, expr := range exprs {
			val := FromAST(expr)
			if val.String() != expr.String() {
				t.Errorf("output mismatch:\nwant:\n%s\nhave:\n%s[END]\n", expr.String(), val.String())
			}
		}
	})
}
