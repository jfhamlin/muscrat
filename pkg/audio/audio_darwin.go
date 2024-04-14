// This file is a modified version of a file from the oto project.

// Copyright 2021 The Oto Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package audio

import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego/objc"
)

const (
	float32SizeInBytes = 4

	bufferCount = 8

	noErr = 0
)

func newAudioQueue(sampleRate, channelCount int, oneBufferSizeInBytes int) (_AudioQueueRef, []_AudioQueueBufferRef, error) {
	desc := _AudioStreamBasicDescription{
		mSampleRate:       float64(sampleRate),
		mFormatID:         uint32(kAudioFormatLinearPCM),
		mFormatFlags:      uint32(kAudioFormatFlagIsFloat),
		mBytesPerPacket:   uint32(channelCount * float32SizeInBytes),
		mFramesPerPacket:  1,
		mBytesPerFrame:    uint32(channelCount * float32SizeInBytes),
		mChannelsPerFrame: uint32(channelCount),
		mBitsPerChannel:   uint32(8 * float32SizeInBytes),
	}

	var audioQueue _AudioQueueRef
	if osstatus := _AudioQueueNewOutput(
		&desc,
		render,
		nil,
		0, //CFRunLoopRef
		0, //CFStringRef
		0,
		&audioQueue); osstatus != noErr {
		return 0, nil, fmt.Errorf("AudioQueueNewFormat with StreamFormat failed: %d", osstatus)
	}

	bufs := make([]_AudioQueueBufferRef, 0, bufferCount)
	for len(bufs) < cap(bufs) {
		var buf _AudioQueueBufferRef
		if osstatus := _AudioQueueAllocateBuffer(audioQueue, uint32(oneBufferSizeInBytes), &buf); osstatus != noErr {
			return 0, nil, fmt.Errorf("AudioQueueAllocateBuffer failed: %d", osstatus)
		}
		buf.mAudioDataByteSize = uint32(oneBufferSizeInBytes)
		bufs = append(bufs, buf)
	}

	return audioQueue, bufs, nil
}

type context struct {
	audioQueue      _AudioQueueRef
	unqueuedBuffers []_AudioQueueBufferRef

	oneBufferSizeInBytes int

	cond *sync.Cond

	input         chan []float32
	lastInput     []float32
	lastInputView []float32
}

// TODO: Convert the error code correctly.
// See https://stackoverflow.com/questions/2196869/how-do-you-convert-an-iphone-osstatus-code-to-something-useful

var theContext *context

func newContext(sampleRate int, channelCount int, bufferSizeInBytes int) (*context, error) {
	// defaultOneBufferSizeInBytes is the default buffer size in bytes.
	//
	// 12288 seems necessary at least on iPod touch (7th) and MacBook Pro 2020.
	// With 48000[Hz] stereo, the maximum delay is (12288*4[buffers] / 4 / 2)[samples] / 48000 [Hz] = 100[ms].
	// '4' is float32 size in bytes. '2' is a number of channels for stereo.
	const defaultOneBufferSizeInBytes = 12288

	var oneBufferSizeInBytes int
	if bufferSizeInBytes != 0 {
		oneBufferSizeInBytes = bufferSizeInBytes
	} else {
		oneBufferSizeInBytes = defaultOneBufferSizeInBytes
	}
	bytesPerSample := channelCount * 4
	oneBufferSizeInBytes = oneBufferSizeInBytes / bytesPerSample * bytesPerSample

	c := &context{
		cond:                 sync.NewCond(&sync.Mutex{}),
		oneBufferSizeInBytes: oneBufferSizeInBytes,
		input:                make(chan []float32, bufferCount),
	}
	theContext = c

	if err := initializeAPI(); err != nil {
		return nil, err
	}

	q, bs, err := newAudioQueue(sampleRate, channelCount, oneBufferSizeInBytes)
	if err != nil {
		return nil, err
	}
	c.audioQueue = q
	c.unqueuedBuffers = bs

	if err := setNotificationHandler(); err != nil {
		return nil, err
	}

	var retryCount int
try:
	if osstatus := _AudioQueueStart(c.audioQueue, nil); osstatus != noErr {
		if osstatus == avAudioSessionErrorCodeCannotStartPlaying && retryCount < 100 {
			// TODO: use sleepTime() after investigating when this error happens.
			time.Sleep(10 * time.Millisecond)
			retryCount++
			goto try
		}
		return nil, fmt.Errorf("AudioQueueStart failed at newContext: %d", osstatus)
	}

	go c.loop()

	return c, nil
}

func (c *context) wait() bool {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	for len(c.unqueuedBuffers) == 0 {
		c.cond.Wait()
	}
	return true
}

func (c *context) loop() {
	for {
		if !c.wait() {
			return
		}
		c.fillBuffer()
	}
}

func (c *context) fillBuffer() {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	buf := c.unqueuedBuffers[len(c.unqueuedBuffers)-1]
	c.unqueuedBuffers = c.unqueuedBuffers[:len(c.unqueuedBuffers)-1]

	inBuf := <-c.input
	if len(inBuf) != int(buf.mAudioDataByteSize/float32SizeInBytes) {
		panic(fmt.Errorf("unexpected input size: %d != %d", len(c.lastInput), buf.mAudioDataByteSize/float32SizeInBytes))
	}

	copy(unsafe.Slice((*float32)(unsafe.Pointer(buf.mAudioData)), buf.mAudioDataByteSize/float32SizeInBytes), inBuf)

	pool.Put(inBuf)

	if osstatus := _AudioQueueEnqueueBuffer(c.audioQueue, buf, 0, nil); osstatus != noErr {
		panic(fmt.Errorf("AudioQueueEnqueueBuffer failed: %d", osstatus))
	}
}

func (c *context) Suspend() error {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	if osstatus := _AudioQueuePause(c.audioQueue); osstatus != noErr {
		return fmt.Errorf("AudioQueuePause failed: %d", osstatus)
	}
	return nil
}

func (c *context) Resume() error {
	c.cond.L.Lock()
	defer c.cond.L.Unlock()

	var retryCount int
try:
	if osstatus := _AudioQueueStart(c.audioQueue, nil); osstatus != noErr {
		if (osstatus == avAudioSessionErrorCodeCannotStartPlaying ||
			osstatus == avAudioSessionErrorCodeCannotInterruptOthers) &&
			retryCount < 30 {
			// It is uncertain that this error is temporary or not. Then let's use exponential-time sleeping.
			time.Sleep(sleepTime(retryCount))
			retryCount++
			goto try
		}
		if osstatus == avAudioSessionErrorCodeSiriIsRecording {
			// As this error should be temporary, it should be OK to use a short time for sleep anytime.
			time.Sleep(10 * time.Millisecond)
			goto try
		}
		return fmt.Errorf("AudioQueueStart failed at Resume: %d", osstatus)
	}
	return nil
}

func render(inUserData unsafe.Pointer, inAQ _AudioQueueRef, inBuffer _AudioQueueBufferRef) {
	theContext.cond.L.Lock()
	defer theContext.cond.L.Unlock()
	theContext.unqueuedBuffers = append(theContext.unqueuedBuffers, inBuffer)
	theContext.cond.Signal()
}

func setGlobalPause(self objc.ID, _cmd objc.SEL, notification objc.ID) {
	theContext.Suspend()
}

func setGlobalResume(self objc.ID, _cmd objc.SEL, notification objc.ID) {
	theContext.Resume()
}

func sleepTime(count int) time.Duration {
	switch count {
	case 0:
		return 10 * time.Millisecond
	case 1:
		return 20 * time.Millisecond
	case 2:
		return 50 * time.Millisecond
	default:
		return 100 * time.Millisecond
	}
}
