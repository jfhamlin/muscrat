package freeverb

// Go translation of revmodel.hpp and revmodel.cpp from the C++
// freeverb library.
//
// Reverb model implementation
//
// C++ version originally written by Jezar at Dreampoint, June 2000
// http://www.dreampoint.co.uk
// This code is public domain

// RevModel is a reverb model.
type RevModel struct {
	gain                float32
	roomsize, roomsize1 float32
	damp, damp1         float32
	wet, wet1, wet2     float32
	dry                 float32
	width               float32
	mode                float32

	// Comb filters
	combL [numcombs]*comb
	combR [numcombs]*comb

	// Allpass filters
	allpassL [numallpasses]*allPass
	allpassR [numallpasses]*allPass
}

func NewRevModel() *RevModel {
	rm := &RevModel{}
	for i := 0; i < numcombs; i++ {
		rm.combL[i] = newComb(combtuningL[i])
		rm.combR[i] = newComb(combtuningR[i])
	}
	for i := 0; i < numallpasses; i++ {
		rm.allpassL[i] = newAllPass(allpasstuningL[i])
		rm.allpassR[i] = newAllPass(allpasstuningR[i])
	}
	rm.SetWet(initialwet)
	rm.SetRoomSize(initialroom)
	rm.SetDry(initialdry)
	rm.SetDamp(initialdamp)
	rm.SetWidth(initialwidth)
	rm.SetMode(initialmode)

	return rm
}

func (rm *RevModel) Mute() {
	if rm.GetMode() >= freezemode {
		return
	}
	for i := 0; i < numcombs; i++ {
		rm.combL[i].mute()
		rm.combR[i].mute()
	}
	for i := 0; i < numallpasses; i++ {
		rm.allpassL[i].mute()
		rm.allpassR[i].mute()
	}
}

func (rm *RevModel) ProcessReplace(inputL, inputR, outputL, outputR []float32, numSamples, skip int) {
	for numSamples > 0 {
		numSamples--

		outL := float32(0)
		outR := float32(0)
		input := (inputL[0] + inputR[0]) * rm.gain

		// Accumulate comb filters in parallel
		for i := 0; i < numcombs; i++ {
			outL += rm.combL[i].process(input)
			outR += rm.combR[i].process(input)
		}

		// Feed through allpasses in series
		for i := 0; i < numallpasses; i++ {
			outL = rm.allpassL[i].process(outL)
			outR = rm.allpassR[i].process(outR)
		}

		// Calculate output REPLACING anything already there
		outputL[0] = outL*rm.wet1 + outR*rm.wet2 + inputL[0]*rm.dry
		outputR[0] = outR*rm.wet1 + outL*rm.wet2 + inputR[0]*rm.dry

		// Increment sample pointers, allowing for interleave (if any)
		inputL = inputL[skip:]
		inputR = inputR[skip:]
		outputL = outputL[skip:]
		outputR = outputR[skip:]
	}
}

func (rm *RevModel) ProcessMix(inputL, inputR, outputL, outputR []float32, numSamples, skip int) {
	for numSamples > 0 {
		numSamples--

		outL := float32(0)
		outR := float32(0)
		input := (inputL[0] + inputR[0]) * rm.gain

		// Accumulate comb filters in parallel
		for i := 0; i < numcombs; i++ {
			outL += rm.combL[i].process(input)
			outR += rm.combR[i].process(input)
		}

		// Feed through allpasses in series
		for i := 0; i < numallpasses; i++ {
			outL = rm.allpassL[i].process(outL)
			outR = rm.allpassR[i].process(outR)
		}

		// Calculate output MIXING with anything already there
		outputL[0] += outL*rm.wet1 + outR*rm.wet2 + inputL[0]*rm.dry
		outputR[0] += outR*rm.wet1 + outL*rm.wet2 + inputR[0]*rm.dry

		// Increment sample pointers, allowing for interleave (if any)
		inputL = inputL[skip:]
		inputR = inputR[skip:]
		outputL = outputL[skip:]
		outputR = outputR[skip:]
	}
}

func (rm *RevModel) update() {
	// Recalculate internal values after parameter change

	rm.wet1 = rm.wet * (rm.width/2 + 0.5)
	rm.wet2 = rm.wet * ((1 - rm.width) / 2)

	if rm.mode >= freezemode {
		rm.roomsize1 = 1
		rm.damp1 = 0
		rm.gain = muted
	} else {
		rm.roomsize1 = rm.roomsize
		rm.damp1 = rm.damp
		rm.gain = fixedgain
	}

	for i := 0; i < numcombs; i++ {
		rm.combL[i].setFeedback(rm.roomsize1)
		rm.combR[i].setFeedback(rm.roomsize1)
	}

	for i := 0; i < numcombs; i++ {
		rm.combL[i].setDamp(rm.damp1)
		rm.combR[i].setDamp(rm.damp1)
	}
}

func (rm *RevModel) SetRoomSize(value float32) {
	rm.roomsize = (value*scaleroom + offsetroom)
	rm.update()
}

func (rm *RevModel) GetRoomSize() float32 {
	return (rm.roomsize - offsetroom) / scaleroom
}

func (rm *RevModel) SetDamp(value float32) {
	rm.damp = value * scaledamp
	rm.update()
}

func (rm *RevModel) GetDamp() float32 {
	return rm.damp / scaledamp
}

func (rm *RevModel) SetWet(value float32) {
	rm.wet = value * scalewet
	rm.update()
}

func (rm *RevModel) GetWet() float32 {
	return rm.wet / scalewet
}

func (rm *RevModel) SetDry(value float32) {
	rm.dry = value * scaledry
}

func (rm *RevModel) GetDry() float32 {
	return rm.dry / scaledry
}

func (rm *RevModel) SetWidth(value float32) {
	rm.width = value
	rm.update()
}

func (rm *RevModel) GetWidth() float32 {
	return rm.width
}

func (rm *RevModel) SetMode(value float32) {
	rm.mode = value
	rm.update()
}

func (rm *RevModel) GetMode() float32 {
	if rm.mode >= freezemode {
		return 1
	} else {
		return 0
	}
}
