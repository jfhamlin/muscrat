package freeverb

import (
	"math"
	"unsafe"
)

// Go translation of denormals.h from the C++ freeverb library.
//
// Macro for killing denormalled numbers
//
// C++ originally written by Jezar at Dreampoint, June 2000
// http://www.dreampoint.co.uk
// This code is public domain

// undenormalise converts denormalised floating point numbers to zero.
func undenormalise(sample *float32) {
	if *(*uint32)(unsafe.Pointer(sample))&0x7f800000 == 0 && *sample != 0 || math.IsNaN(float64(*sample)) {
		*sample = 0
	}
}
