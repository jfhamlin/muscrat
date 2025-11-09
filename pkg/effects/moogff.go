/*
"MoogFF" - Moog VCF digital implementation.
As described in the paper entitled
"Preserving the Digital Structure of the Moog VCF"
by Federico Fontana
appeared in the Proc. ICMC07, Copenhagen, 25-31 August 2007

Original Java code Copyright F. Fontana - August 2007
federico.fontana@univr.it

Ported to C++ for SuperCollider by Dan Stowell - August 2007
http://www.mcld.co.uk/

Ported to Go for Muscrat by Claude Code - November 2025

    This program is free software; you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation; either version 2 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program; if not, write to the Free Software
    Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301  USA
*/

package effects

import (
	"context"
	"math"

	"github.com/jfhamlin/muscrat/pkg/conf"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

func NewMoogFF() ugen.UGen {
	sampleRate := float64(conf.SampleRate)
	sampleDur := 1.0 / sampleRate

	var s1, s2, s3, s4 float64 // Filter states
	var a1, b0 float64         // Filter coefficients
	var prevFreq float64
	var coefficientsComputed bool

	computeCoefficients := func(freq float64) {
		T := sampleDur
		wcD := 2.0 * math.Tan(T*math.Pi*freq) * sampleRate
		if wcD < 0 {
			wcD = 0 // Protect against negative cutoff freq
		}
		TwcD := T * wcD
		b0 = TwcD / (TwcD + 2.0)
		a1 = (TwcD - 2.0) / (TwcD + 2.0)
	}

	return ugen.UGenFunc(func(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
		in := cfg.InputSamples["in"]
		freq := cfg.InputSamples["freq"]
		gain := cfg.InputSamples["gain"]
		reset := cfg.InputSamples["reset"]

		_ = in[len(out)-1]
		_ = freq[len(out)-1]
		_ = gain[len(out)-1]
		_ = reset[len(out)-1]

		for i := range out {
			// Reset filter state if requested
			if reset[i] > 0 {
				s1, s2, s3, s4 = 0, 0, 0, 0
			}

			// Update filter coefficients if frequency changes
			if !coefficientsComputed || freq[i] != prevFreq {
				computeCoefficients(freq[i])
				prevFreq = freq[i]
				coefficientsComputed = true
			}

			// Clamp gain (resonance) to [0, 4]
			k := gain[i]
			if k < 0 {
				k = 0
			} else if k > 4 {
				k = 4
			}

			// Compute output
			o := s4 + b0*(s3+b0*(s2+b0*s1))
			ins := in[i]
			b0_4 := b0 * b0 * b0 * b0
			outs := (b0_4*ins + o) / (1.0 + b0_4*k)
			out[i] = outs

			u := ins - k*outs

			// Update 1st order filter states
			past := u
			future := b0*past + s1
			s1 = b0*past - a1*future

			past = future
			future = b0*past + s2
			s2 = b0*past - a1*future

			past = future
			future = b0*past + s3
			s3 = b0*past - a1*future

			s4 = b0*future - a1*outs

			// Zap gremlins
			s1 = ugen.ZapGremlins(s1)
			s2 = ugen.ZapGremlins(s2)
			s3 = ugen.ZapGremlins(s3)
			s4 = ugen.ZapGremlins(s4)
		}
	})
}
