package freeverb

// Go translation of tuning.h from the C++ freeverb library.
//
// Reverb model tuning values
//
// Written by Jezar at Dreampoint, June 2000
// http://www.dreampoint.co.uk
// This code is public domain

const (
	numcombs     = 8
	numallpasses = 4
	muted        = 0.0
	fixedgain    = 0.015
	scalewet     = 3.0
	scaledry     = 2.0
	scaledamp    = 0.4
	scaleroom    = 0.28
	offsetroom   = 0.7
	initialroom  = 0.5
	initialdamp  = 0.5
	initialwet   = 1.0 / scalewet
	initialdry   = 0
	initialwidth = 1.0
	initialmode  = 0.0
	freezemode   = 0.5
	stereospread = 23

	// These values assume 44.1KHz sample rate
	// they will probably be OK for 48KHz sample rate
	// but would need scaling for 96KHz (or other) sample rates.
	// The values were obtained by listening tests.
	combtuningL1    = 1116
	combtuningR1    = 1116 + stereospread
	combtuningL2    = 1188
	combtuningR2    = 1188 + stereospread
	combtuningL3    = 1277
	combtuningR3    = 1277 + stereospread
	combtuningL4    = 1356
	combtuningR4    = 1356 + stereospread
	combtuningL5    = 1422
	combtuningR5    = 1422 + stereospread
	combtuningL6    = 1491
	combtuningR6    = 1491 + stereospread
	combtuningL7    = 1557
	combtuningR7    = 1557 + stereospread
	combtuningL8    = 1617
	combtuningR8    = 1617 + stereospread
	allpasstuningL1 = 556
	allpasstuningR1 = 556 + stereospread
	allpasstuningL2 = 441
	allpasstuningR2 = 441 + stereospread
	allpasstuningL3 = 341
	allpasstuningR3 = 341 + stereospread
	allpasstuningL4 = 225
	allpasstuningR4 = 225 + stereospread
)

var (
	// Slices of comb filter tuning values for left and right filters.
	combtuningL = []int{
		combtuningL1, combtuningL2, combtuningL3, combtuningL4,
		combtuningL5, combtuningL6, combtuningL7, combtuningL8,
	}
	combtuningR = []int{
		combtuningR1, combtuningR2, combtuningR3, combtuningR4,
		combtuningR5, combtuningR6, combtuningR7, combtuningR8,
	}

	// Slices of allpass filter tuning values for left and right filters.
	allpasstuningL = []int{
		allpasstuningL1, allpasstuningL2, allpasstuningL3, allpasstuningL4,
	}
	allpasstuningR = []int{
		allpasstuningR1, allpasstuningR2, allpasstuningR3, allpasstuningR4,
	}
)
