package freeverb

// Go translation of allpass.hpp and allpass.cpp from the C++ freeverb
// library.
//
// Allpass filter declaration
//
// C++ originally written by Jezar at Dreampoint, June 2000
// http://www.dreampoint.co.uk
// This code is public domain

// allPass is an allpass filter.
type allPass struct {
	feedback float32
	buffer   []float32
	bufidx   int
}

// newAllPass creates a new allpass filter.
func newAllPass(size int) *allPass {
	return &allPass{
		buffer:   make([]float32, size),
		feedback: 0.5,
	}
}

func (a *allPass) process(input float32) float32 {
	bufout := a.buffer[a.bufidx]
	undenormalise(&bufout)

	output := -input + bufout
	a.buffer[a.bufidx] = input + (bufout * a.feedback)

	a.bufidx++
	if a.bufidx >= len(a.buffer) {
		a.bufidx = 0
	}

	return output
}

func (a *allPass) mute() {
	for i := range a.buffer {
		a.buffer[i] = 0
	}
}

func (a *allPass) setFeedback(val float32) {
	a.feedback = val
}

func (a *allPass) getFeedback() float32 {
	return a.feedback
}
