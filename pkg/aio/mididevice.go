package aio

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/jfhamlin/muscrat/pkg/ugen"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type (
	MIDIDevice interface {
		Notes(idx int) ugen.UGen
		Gates(idx int) ugen.UGen
	}

	SoftwareKeyboard struct {
		Name string

		notes  []float64
		gates  []float64
		counts []int

		counter int

		cancel func()

		mtx sync.Mutex
	}

	softwareKeyboardNotes struct {
		*SoftwareKeyboard
		idx int
	}

	softwareKeyboardGate struct {
		*SoftwareKeyboard
		idx int
	}

	MIDIDeviceOption func(*midiDeviceOptions)

	midiDeviceOptions struct {
		voices int
	}
)

func WithVoices(voices int) MIDIDeviceOption {
	return func(o *midiDeviceOptions) {
		o.voices = voices
	}
}

func NewSoftwareKeyboard(name string, opts ...MIDIDeviceOption) *SoftwareKeyboard {
	o := &midiDeviceOptions{
		voices: 1,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.voices < 1 {
		o.voices = 1
	}

	kb := &SoftwareKeyboard{
		Name:   name,
		notes:  make([]float64, o.voices),
		gates:  make([]float64, o.voices),
		counts: make([]int, o.voices),
	}
	return kb
}

func (s *SoftwareKeyboard) Start(ctx context.Context) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.cancel == nil {
		s.cancel = runtime.EventsOn(ctx, "midi-event", func(e ...interface{}) {
			event := e[0].(map[string]interface{})
			typ := event["type"].(string)
			note := event["midiNumber"].(float64)
			freq := math.Pow(2, (note-69)/12) * 440
			switch typ {
			case "noteOn":
				// pick the oldest unused voice
				selectedIdx := -1
				selectedCount := math.MaxInt
				for i := range s.notes {
					if s.gates[i] == 0 && s.counts[i] < selectedCount {
						selectedIdx = i
						selectedCount = s.counts[i]
					}
				}
				if selectedIdx >= 0 {
					s.notes[selectedIdx] = freq
					s.gates[selectedIdx] = 1
					s.counts[selectedIdx] = s.counter
					s.counter++
				}
				// TODO: reassign oldest note
			case "noteOff":
				for i := range s.notes {
					if s.notes[i] == freq {
						s.gates[i] = 0
					}
				}
			}
		})
	}

	return nil
}

func (s *SoftwareKeyboard) Stop(ctx context.Context) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	return nil
}

func (s *SoftwareKeyboard) Notes(idx int) ugen.UGen {
	return &softwareKeyboardNotes{SoftwareKeyboard: s, idx: idx}
}

func (s *SoftwareKeyboard) Gates(idx int) ugen.UGen {
	return &softwareKeyboardGate{SoftwareKeyboard: s, idx: idx}
}

func (s *softwareKeyboardNotes) GenerateSamples(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
	res := make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = s.notes[s.idx]
	}
	return res
}

func (s *softwareKeyboardGate) GenerateSamples(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
	res := make([]float64, n)
	for i := 0; i < n; i++ {
		res[i] = s.gates[s.idx]
	}
	return res
}

func NewMIDIDevice(name, typ string, opts ...MIDIDeviceOption) MIDIDevice {
	switch typ {
	case "sw-keyboard":
		return NewSoftwareKeyboard(name, opts...)
	}
	panic(fmt.Sprintf("unknown MIDI device type: %s", typ))
}
