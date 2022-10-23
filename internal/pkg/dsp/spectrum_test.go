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
