package ugen

import (
	"context"
	"testing"
)

func TestMovingAverage(t *testing.T) {
	type testCase struct {
		desc         string
		sampleRateHz int
		maxDurSecs   float64
		in           []float64
		dur          []float64
		out          []float64
	}

	testCases := []testCase{
		{
			desc:         "one sample",
			sampleRateHz: 1,
			maxDurSecs:   0.1,
			in:           []float64{1, 2, 3, 4, 5},
			dur:          []float64{0.1, 0.1, 0.1, 0.1, 0.1},
			// only one sample, so the moving average is the same as the input
			out: []float64{1, 2, 3, 4, 5},
		},
		{
			desc:         "two samples",
			sampleRateHz: 1,
			maxDurSecs:   2, // a duration of 2 seconds will result in a
			// moving average of 2 samples
			in:  []float64{1, 2, 3, 4, 5},
			dur: []float64{2, 2, 2, 2, 2},
			// moving average of 2 samples
			out: []float64{1, 1.5, 2.5, 3.5, 4.5},
		},
		{
			desc:         "durs > maxDurSecs",
			sampleRateHz: 1,
			maxDurSecs:   2,
			in:           []float64{1, 2, 3, 4, 5},
			dur:          []float64{3, 3, 3, 3, 3},
			// moving average of 2 samples
			out: []float64{1, 1.5, 2.5, 3.5, 4.5},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			t.Parallel()

			ma := NewMovingAverage(tc.maxDurSecs)
			cfg := SampleConfig{
				SampleRateHz: tc.sampleRateHz,
				InputSamples: map[string][]float64{
					"in":  tc.in,
					"dur": tc.dur,
				},
			}
			out := make([]float64, len(tc.out))
			ma.Gen(context.Background(), cfg, out)

			for i := range out {
				if out[i] != tc.out[i] {
					t.Errorf("expected %f, got %f", tc.out[i], out[i])
				}
			}
		})
	}
}
