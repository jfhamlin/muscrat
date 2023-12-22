package aio

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/jfhamlin/muscrat/pkg/pubsub"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

type (
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
)

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
		s.cancel = pubsub.Subscribe("midi-event", func(evt string, data any) {
			event := data.(map[string]interface{})
			fmt.Printf("event: %v\n", event)
			typ := event["type"].(string)
			note := float64(event["midiNumber"].(int))
			fmt.Printf("note: %v\n", note)
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
					s.notes[selectedIdx] = note
					s.gates[selectedIdx] = 1
					s.counts[selectedIdx] = s.counter
					s.counter++
				}
				// TODO: reassign oldest note
			case "noteOff":
				for i := range s.notes {
					if s.notes[i] == note {
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

func (s *SoftwareKeyboard) Note(idx int) ugen.UGen {
	return &softwareKeyboardNotes{SoftwareKeyboard: s, idx: idx}
}

func (s *SoftwareKeyboard) Gate(idx int) ugen.UGen {
	return &softwareKeyboardGate{SoftwareKeyboard: s, idx: idx}
}

func (s *SoftwareKeyboard) Velocity(idx int) ugen.UGen {
	return nil
}

func (s *SoftwareKeyboard) PitchBend() ugen.UGen {
	return nil
}

func (s *softwareKeyboardNotes) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := range out {
		out[i] = s.notes[s.idx]
	}
}

func (s *softwareKeyboardGate) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := range out {
		out[i] = s.gates[s.idx]
	}
}
