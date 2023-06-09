package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

type (
	WaveFolder struct{}
)

func NewWaveFolder() *WaveFolder {
	return &WaveFolder{}
}

func (w *WaveFolder) GenerateSamples(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
	in := cfg.InputSamples["in"]
	los := cfg.InputSamples["lo"]
	his := cfg.InputSamples["hi"]

	res := make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = in[i]

		x := in[i]
		lo := -1.0
		hi := 1.0
		if len(los) > 0 {
			lo = los[i]
		}
		if len(his) > 0 {
			hi = math.Max(his[i], lo)
		}

		// transform x such -1 = lo, 0 = (lo + hi) / 2, 1 = hi
		mid := (lo + hi) / 2
		x = (x - mid) / (hi - mid)

		// if x > 1, it is reflected back.
		// it may be reflected back multiple times.
		// the same is true for x < -1.
		if x > 1 {
			floor := math.Floor(x)
			rem := x - floor
			switch int64(floor) % 4 {
			case 0:
				res[i] = rem
			case 1:
				res[i] = 1 - rem
			case 2:
				res[i] = -rem
			case 3:
				res[i] = rem - 1
			}
		} else if x < -1 {
			ciel := math.Ceil(x)
			rem := ciel - x
			switch int64(math.Abs(ciel)) % 4 {
			case 0:
				res[i] = -rem
			case 1:
				res[i] = rem - 1
			case 2:
				res[i] = rem
			case 3:
				res[i] = 1 - rem
			}
		} else {
			res[i] = x
		}
		// transform res[i] back to -1 = lo, 0 = (lo + hi) / 2, 1 = hi
		res[i] = res[i]*(hi-mid) + mid
	}
	return res
}
