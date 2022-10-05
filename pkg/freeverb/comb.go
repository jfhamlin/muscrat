package freeverb

// Go translation of comb.hpp and comb.cpp from the C++ freeverb
// library.
//
// Comb filter class declaration
//
// Originally written by Jezar at Dreampoint, June 2000
// http://www.dreampoint.co.uk
// This code is public domain

// comb is a comb filter.
type comb struct {
	feedback    float32
	filterstore float32
	damp1       float32
	damp2       float32
	buffer      []float32
	bufidx      int
}

// newComb creates a new comb filter.
func newComb(size int) *comb {
	return &comb{
		buffer: make([]float32, size),
	}
}

func (c *comb) process(input float32) float32 {
	output := c.buffer[c.bufidx]
	undenormalise(&output)

	c.filterstore = (output * c.damp2) + (c.filterstore * c.damp1)
	undenormalise(&c.filterstore)

	c.buffer[c.bufidx] = input + (c.filterstore * c.feedback)

	c.bufidx++
	if c.bufidx >= len(c.buffer) {
		c.bufidx = 0
	}

	return output
}

func (c *comb) mute() {
	for i := range c.buffer {
		c.buffer[i] = 0
	}
}

func (c *comb) setDamp(val float32) {
	c.damp1 = val
	c.damp2 = 1 - val
}

func (c *comb) getDamp() float32 {
	return c.damp1
}

func (c *comb) setFeedback(val float32) {
	c.feedback = val
}
