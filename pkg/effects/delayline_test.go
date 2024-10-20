package effects

import (
	"fmt"
	"math"
	"testing"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func TestDelayLine(t *testing.T) {
	type testCase struct {
		sampleRate int
		maxDelay   float64
		delay      float64
	}

	delayTimeFromSamples := func(sampleRate int, delay float64) float64 {
		return delay / float64(sampleRate)
	}

	tests := []testCase{
		{44100, 1.0, 0},
		// 1.5 samples delay
		{44100, 1.0, delayTimeFromSamples(44100, 1.5)},
		// half a sample delay
		{44100, 1.0, delayTimeFromSamples(44100, 0.5)},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(fmt.Sprintf("sampleRate=%d,maxDelay=%f,delay=%f", tc.sampleRate, tc.maxDelay, tc.delay), func(t *testing.T) {
			for _, interp := range []int{0, 1, 3} {
				interp := interp
				t.Run(fmt.Sprintf("interp=%d", interp), func(t *testing.T) {
					t.Parallel()
					dl := NewDelayLine(tc.sampleRate, tc.maxDelay)
					if dl == nil {
						t.Fatal("NewDelayLine returned nil")
					}

					dl.SetDelaySeconds(tc.delay)

					fracPart := tc.delay * float64(tc.sampleRate)
					intPart := int(fracPart + 0.5)
					fracPart = math.Mod(fracPart, 1.0)

					expectZeros := intPart
					buf := make([]float64, expectZeros+100+1)

					for i := 0; i < 100; i++ {
						dl.WriteSample(float64(i))
						buf[i+expectZeros] = float64(i)
					}

					readSample := dl.ReadSampleN
					switch interp {
					case 0:
						readSample = dl.ReadSampleN
					case 1:
						readSample = dl.ReadSampleL
					case 3:
						readSample = dl.ReadSampleC
					default:
						t.Fatalf("unknown interp %d", interp)
					}

					for i := 0; i < 100; i++ {
						sample := readSample()
						switch interp {
						case 0:
							if sample != buf[i] {
								t.Fatalf("expected %f, got %f", buf[i], sample)
							}
						case 1:
							x0 := buf[i]
							x1 := buf[(i+1)%len(buf)]
							expected := x0 + fracPart*(x1-x0)
							if sample != expected {
								t.Fatalf("expected %f, got %f", buf[i]+fracPart*(buf[i+1]-buf[i]), sample)
							}
						case 3:
							var x0 float64
							if i == 0 {
								x0 = buf[len(buf)-1]
							} else {
								x0 = buf[(i-1)%len(buf)]
							}
							x1 := buf[i]
							x2 := buf[(i+1)%len(buf)]
							x3 := buf[(i+2)%len(buf)]

							expected := ugen.CubInterp(fracPart, x0, x1, x2, x3)
							if sample != expected {
								t.Fatalf("expected %f, got %f", expected, sample)
							}
						}
					}
				})
			}
		})
	}
}
