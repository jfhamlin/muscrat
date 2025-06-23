package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

// PitchShift Implementation Plan:
//
// This implements a granular pitch shifter based on SuperCollider's algorithm.
// The algorithm works as follows:
//
// 1. Input audio is stored in a circular delay buffer
// 2. Four grains read from the buffer at different positions
// 3. Each grain uses a triangular window (linear ramp up/down)
// 4. Grains are triggered every framesize/4 samples (75% overlap)
// 5. Pitch shifting is achieved by adjusting grain playback speed:
//    - pitchRatio > 1: grains play faster (higher pitch)
//    - pitchRatio < 1: grains play slower (lower pitch)
//
// Key parameters:
// - windowSize: size of each grain window in seconds
// - pitchRatio: pitch shift factor (0.5 = down octave, 2.0 = up octave)
// - pitchDispersion: random pitch variation for chorus effects
// - timeDispersion: random time variation for smearing effects
//
// The implementation follows SuperCollider's approach:
// - Circular buffer sized to windowSize * 3 + overhead
// - Triangular windows using linear ramps
// - Counter-based grain triggering (not phase-based)
// - Four overlapping grains with staggered start times
//
// This provides:
// - Low latency (approximately one window size)
// - Good quality for moderate pitch shifts
// - Optional chorus/ensemble effects with dispersion
// - Artifact-free output through proper overlap

const (
	defaultWindowSize = 0.1 // 100ms default window
	maxPitchRatio     = 4.0 // 2 octaves up
)

// NewPitchShift creates a new pitch shift effect based on SuperCollider's implementation.
// The effect uses granular synthesis with triangular windows to shift the pitch of the input signal.
func NewPitchShift(opts ...ugen.Option) ugen.UGen {
	o := ugen.DefaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	sampleRate := float64(conf.SampleRate)
	minWindowSize := 3.0 / sampleRate // 3 samples minimum

	// State variables - following SuperCollider's structure
	var (
		buffer     []float64
		bufferMask int
		writePos   int

		// Grain parameters - 4 grains with individual ramps and read positions
		dsamp1, dsamp2, dsamp3, dsamp4                     float64
		dsamp1Slope, dsamp2Slope, dsamp3Slope, dsamp4Slope float64
		ramp1, ramp2, ramp3, ramp4                         float64
		ramp1Slope, ramp2Slope, ramp3Slope, ramp4Slope     float64

		// Control variables
		counter   int
		stage     int
		framesize int
		slope     float64

		// Random state for dispersion
		randState uint32 = 1
	)

	// Simple random number generator
	frand := func() float64 {
		randState = randState*1664525 + 1013904223
		return float64(randState) / 4294967296.0
	}

	frand2 := func() float64 {
		return frand()*2.0 - 1.0
	}

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		pitchRatios := cfg.InputSamples["pitchRatio"]
		windowSizes := cfg.InputSamples["windowSize"]
		pitchDispersions := cfg.InputSamples["pitchDispersion"]
		timeDispersions := cfg.InputSamples["timeDispersion"]

		// Ensure we have valid inputs
		_ = in[len(out)-1]

		// Initialize on first run
		if buffer == nil {
			// Get initial window size
			windowSize := defaultWindowSize
			if len(windowSizes) > 0 {
				windowSize = math.Max(minWindowSize, math.Min(1.0, windowSizes[0]))
			}

			// Calculate buffer size like SuperCollider: winsize * 3 + 3 + overhead
			delaybufsize := int(math.Ceil(windowSize*sampleRate*3.0 + 3.0))
			delaybufsize += len(out) // Add block size
			delaybufsize = ugen.NextPowerOf2(delaybufsize)

			buffer = make([]float64, delaybufsize)
			bufferMask = delaybufsize - 1

			// Initialize frame size and slope
			framesize = (int(windowSize*sampleRate) + 2) &^ 3 // Round to multiple of 4
			slope = 2.0 / float64(framesize)

			// Initialize state like SuperCollider
			stage = 3
			counter = framesize >> 2
			ramp1 = 0.5
			ramp2 = 1.0
			ramp3 = 0.5
			ramp4 = 0.0

			// Set initial slopes
			ramp1Slope = -slope
			ramp2Slope = -slope
			ramp3Slope = slope
			ramp4Slope = slope
		}

		for i := range out {
			// Get parameters
			pitchRatio := 1.0
			if len(pitchRatios) > i {
				pitchRatio = math.Max(0.0, math.Min(maxPitchRatio, pitchRatios[i]))
			}

			windowSize := defaultWindowSize
			if len(windowSizes) > i {
				windowSize = math.Max(minWindowSize, math.Min(1.0, windowSizes[i]))
			}

			pitchDispersion := 0.0
			if len(pitchDispersions) > i {
				pitchDispersion = math.Max(0.0, math.Min(1.0, pitchDispersions[i]))
			}

			timeDispersion := 0.0
			if len(timeDispersions) > i {
				timeDispersion = math.Max(0.0, math.Min(windowSize, timeDispersions[i])) * sampleRate
			}

			// Check if we need to start a new grain
			if counter <= 0 {
				counter = framesize >> 2
				stage = (stage + 1) & 3

				// Calculate pitch ratio with dispersion
				dispPitchRatio := pitchRatio
				if pitchDispersion != 0.0 {
					dispPitchRatio += pitchDispersion * frand2()
				}
				dispPitchRatio = math.Max(0.0, math.Min(4.0, dispPitchRatio))

				// Calculate grain parameters
				pitchRatio1 := dispPitchRatio - 1.0
				sampSlope := -pitchRatio1
				startPos := 2.0
				if pitchRatio1 >= 0.0 {
					startPos = float64(framesize)*pitchRatio1 + 2.0
				}
				startPos += timeDispersion * frand()

				// Update parameters based on stage
				switch stage {
				case 0:
					dsamp1Slope = sampSlope
					dsamp1 = startPos
					ramp1 = 0.0
					ramp1Slope = slope
					ramp3Slope = -slope
				case 1:
					dsamp2Slope = sampSlope
					dsamp2 = startPos
					ramp2 = 0.0
					ramp2Slope = slope
					ramp4Slope = -slope
				case 2:
					dsamp3Slope = sampSlope
					dsamp3 = startPos
					ramp3 = 0.0
					ramp3Slope = slope
					ramp1Slope = -slope
				case 3:
					dsamp4Slope = sampSlope
					dsamp4 = startPos
					ramp4 = 0.0
					ramp4Slope = slope
					ramp2Slope = -slope
				}
			}

			// Write input to buffer
			buffer[writePos] = in[i]
			writePos = (writePos + 1) & bufferMask

			// Process grains
			value := 0.0

			// Grain 1
			dsamp1 += dsamp1Slope
			idsamp := int(dsamp1)
			frac := dsamp1 - float64(idsamp)
			irdphase := (writePos - idsamp) & bufferMask
			irdphaseb := (irdphase - 1) & bufferMask
			d1 := buffer[irdphase]
			d2 := buffer[irdphaseb]
			value += (d1 + frac*(d2-d1)) * ramp1
			ramp1 += ramp1Slope

			// Grain 2
			dsamp2 += dsamp2Slope
			idsamp = int(dsamp2)
			frac = dsamp2 - float64(idsamp)
			irdphase = (writePos - idsamp) & bufferMask
			irdphaseb = (irdphase - 1) & bufferMask
			d1 = buffer[irdphase]
			d2 = buffer[irdphaseb]
			value += (d1 + frac*(d2-d1)) * ramp2
			ramp2 += ramp2Slope

			// Grain 3
			dsamp3 += dsamp3Slope
			idsamp = int(dsamp3)
			frac = dsamp3 - float64(idsamp)
			irdphase = (writePos - idsamp) & bufferMask
			irdphaseb = (irdphase - 1) & bufferMask
			d1 = buffer[irdphase]
			d2 = buffer[irdphaseb]
			value += (d1 + frac*(d2-d1)) * ramp3
			ramp3 += ramp3Slope

			// Grain 4
			dsamp4 += dsamp4Slope
			idsamp = int(dsamp4)
			frac = dsamp4 - float64(idsamp)
			irdphase = (writePos - idsamp) & bufferMask
			irdphaseb = (irdphase - 1) & bufferMask
			d1 = buffer[irdphase]
			d2 = buffer[irdphaseb]
			value += (d1 + frac*(d2-d1)) * ramp4
			ramp4 += ramp4Slope

			// Output with normalization
			out[i] = value * 0.5

			// Decrement counter
			counter--

			// Zap denormals
			out[i] = ugen.ZapGremlins(out[i])
		}
	})
}
