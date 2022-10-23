package dsp

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFFTFreqs(t *testing.T) {
	type testCase struct {
		sampleRate float64
		n          int
		want       []float64
	}
	testCases := []testCase{
		{sampleRate: 1, n: 1, want: []float64{0}},
		{sampleRate: 1, n: 2, want: []float64{0, 0.5}},
		{sampleRate: 10, n: 16, want: []float64{0, 0.625, 1.25, 1.875, 2.5, 3.125, 3.75, 4.375, 5}},
		{sampleRate: 100, n: 100}, // should panic
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%vHz %v samples", tc.sampleRate, tc.n), func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r != nil {
					if tc.want != nil {
						t.Errorf("unexpected panic: %v", r)
					}
				}
			}()
			got := FFTFreqs(tc.sampleRate, tc.n)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLogRange(t *testing.T) {
	type testCase struct {
		min, max float64
		n        int
		want     []float64
	}
	testCases := []testCase{
		{min: 1, max: 1, n: 1, want: []float64{1}},
		{min: 1, max: 2, n: 2, want: []float64{1, 2}},
		{min: 1, max: 10, n: 10, want: []float64{1, 1.2589254117941673, 1.5848931924611136, 1.9952623149688795, 2.51188643150958, 3.1622776601683795, 3.9810717055349722, 5.011872336272722, 6.309573444801933, 7.943282347242816}},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%v-%v %v samples", tc.min, tc.max, tc.n), func(t *testing.T) {
			t.Parallel()
			got := LogRange(tc.min, tc.max, tc.n)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
