package aio

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sync"
	"sync/atomic"

	"github.com/jfhamlin/muscrat/pkg/pubsub"
	"github.com/jfhamlin/muscrat/pkg/ugen"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type (
	MIDIDevice interface {
		Note(voice int) ugen.UGen
		Gate(voice int) ugen.UGen
		Velocity(voice int) ugen.UGen

		PitchBend() ugen.UGen
		Control() ugen.UGen
	}

	MIDIDeviceOption func(*midiDeviceOptions)

	midiDeviceOptions struct {
		deviceID          int
		deviceNamePattern *regexp.Regexp
		channel           int
		controller        int

		// for polyphonic note events
		voices int

		defaultValue float64
	}

	MIDIEnvelope struct {
		DeviceID   int
		DeviceName string
		Message    midi.Message
	}
)

var (
	// all ports we're listening on
	// NB: once we begin listening on a port, we don't stop.
	// TODO: reconsider this
	midiPorts    []drivers.In
	midiPortsMtx sync.Mutex

	// map from port number to map of controller number to last value
	lastMidiPortControllerValues    = make(map[int]map[int]int32)
	lastMidiPortControllerValuesMtx sync.RWMutex
)

func WithDeviceID(id int) MIDIDeviceOption {
	return func(o *midiDeviceOptions) {
		o.deviceID = id
	}
}

func WithDeviceName(name string) MIDIDeviceOption {
	return func(o *midiDeviceOptions) {
		o.deviceNamePattern = regexp.MustCompile(name)
	}
}

func WithChannel(channel int) MIDIDeviceOption {
	return func(o *midiDeviceOptions) {
		o.channel = channel
	}
}

func WithController(controller int) MIDIDeviceOption {
	return func(o *midiDeviceOptions) {
		o.controller = controller
	}
}

func WithVoices(voices int) MIDIDeviceOption {
	return func(o *midiDeviceOptions) {
		o.voices = voices
	}
}

func WithDefaultValue(value float64) MIDIDeviceOption {
	return func(o *midiDeviceOptions) {
		o.defaultValue = value
	}
}

func findAndListenToMIDIPort(ctx context.Context, id int, namePattern *regexp.Regexp) (int, error) {
	midiPortsMtx.Lock()
	defer midiPortsMtx.Unlock()

	for _, port := range midiPorts {
		if port.Number() == id {
			return id, nil
		}
		if namePattern != nil && namePattern.MatchString(port.String()) {
			return port.Number(), nil
		}
	}

	var found drivers.In
	for _, port := range midi.GetInPorts() {
		if port.Number() == id {
			found = port
			break
		}
		if namePattern != nil && namePattern.MatchString(port.String()) {
			found = port
			break
		}
	}
	if found == nil {
		var name string
		if namePattern != nil {
			name = namePattern.String()
		}
		return 0, fmt.Errorf("no MIDI port found with id %d or name matching %q", id, name)
	}

	portID := found.Number()

	_, err := midi.ListenTo(found, func(msg midi.Message, timestampms int32) {
		if msg.Type() == midi.ControlChangeMsg { // store last value for each controller
			var channel, controller, value uint8
			msg.GetControlChange(&channel, &controller, &value)
			lastMidiPortControllerValuesMtx.Lock()
			if _, ok := lastMidiPortControllerValues[portID]; !ok {
				lastMidiPortControllerValues[portID] = make(map[int]int32)
			}
			lastMidiPortControllerValues[portID][int(controller)] = int32(value)
			lastMidiPortControllerValuesMtx.Unlock()
		}
		pubsub.Publish("midi", &MIDIEnvelope{
			DeviceID:   found.Number(),
			DeviceName: found.String(),
			Message:    msg,
		})
	})
	if err != nil {
		return 0, err
	}
	midiPorts = append(midiPorts, found)
	return found.Number(), nil
}

////////////////////////////////////////////////////////////////////////////////

type (
	Keyboard struct {
		Name string

		options *midiDeviceOptions

		voices     []atomic.Value
		controller atomic.Int32

		counter int

		cancel func()

		mtx sync.Mutex
	}

	voice struct {
		note float64
		gate float64
		// used to track the "age" of the note
		count int
	}

	KeyboardNotes struct {
		*Keyboard
		voice int
	}

	KeyboardGate struct {
		*Keyboard
		voice int
	}

	MIDIControl struct {
		*Keyboard
	}
)

func NewMIDIInputDevice(name string, opts ...MIDIDeviceOption) *Keyboard {
	o := &midiDeviceOptions{
		voices: 1,
	}
	for _, opt := range opts {
		opt(o)
	}
	if o.voices < 1 {
		o.voices = 1
	}

	kb := &Keyboard{
		Name:    name,
		options: o,
	}
	kb.voices = make([]atomic.Value, o.voices)
	for i := range kb.voices {
		kb.voices[i].Store(voice{})
	}

	value := int32(127 * o.defaultValue)
	lastMidiPortControllerValuesMtx.RLock()
	defer lastMidiPortControllerValuesMtx.RUnlock()
	if lastValues, ok := lastMidiPortControllerValues[o.deviceID]; ok {
		if lastValue, ok := lastValues[o.controller]; ok {
			value = lastValue
		}
	}

	kb.controller.Store(value)
	return kb
}

func (s *Keyboard) Start(ctx context.Context) error {
	id, err := findAndListenToMIDIPort(ctx, s.options.deviceID, s.options.deviceNamePattern)
	if err != nil {
		// TODO: don't error out here, but allow the user to fix the
		// mapping in the UI.
		return err
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.cancel == nil {
		s.cancel = pubsub.Subscribe("midi", func(eventName string, e any) {
			evt := e.(*MIDIEnvelope)
			msg := evt.Message
			if evt.DeviceID != id {
				return
			}

			switch msg.Type() {
			case midi.NoteOnMsg:
				var channel, key, velocity uint8
				msg.GetNoteOn(&channel, &key, &velocity)
				if channel != uint8(s.options.channel) {
					return
				}

				// pick the oldest unused voice
				selectedIdx := -1
				selectedCount := math.MaxInt
				for i := range s.voices {
					v := s.voices[i].Load().(voice)
					if v.gate == 0 && v.count < selectedCount {
						selectedIdx = i
						selectedCount = v.count
					}
				}
				if selectedIdx >= 0 {
					s.voices[selectedIdx].Store(voice{
						note:  float64(key),
						gate:  1,
						count: s.counter,
					})
					s.counter++
				}
				// TODO: reassign oldest note
			case midi.NoteOffMsg:
				var channel, key, velocity uint8
				msg.GetNoteOff(&channel, &key, &velocity)
				if channel != uint8(s.options.channel) {
					return
				}

				for i := range s.voices {
					v := s.voices[i].Load().(voice)
					if v.note == float64(key) {
						v.gate = 0
						s.voices[i].Store(v)
					}
				}
			case midi.ControlChangeMsg:
				var channel, controller, value uint8
				msg.GetControlChange(&channel, &controller, &value)
				if channel != uint8(s.options.channel) || controller != uint8(s.options.controller) {
					return
				}
				s.controller.Store(int32(value))
			}
		})
	}

	return nil
}

func (s *Keyboard) Stop(ctx context.Context) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	return nil
}

func (s *Keyboard) Note(voice int) ugen.UGen {
	return &KeyboardNotes{Keyboard: s, voice: voice}
}

func (s *Keyboard) Gate(voice int) ugen.UGen {
	return &KeyboardGate{Keyboard: s, voice: voice}
}

func (s *Keyboard) Velocity(voice int) ugen.UGen {
	return nil
}

func (s *Keyboard) PitchBend() ugen.UGen {
	return nil
}

func (s *Keyboard) Control() ugen.UGen {
	return &MIDIControl{Keyboard: s}
}

func (s *KeyboardNotes) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := range out {
		v := s.voices[s.voice].Load().(voice)
		out[i] = v.note
	}
}

func (s *KeyboardGate) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := range out {
		v := s.voices[s.voice].Load().(voice)
		out[i] = v.gate
	}
}

func (c *MIDIControl) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := range out {
		out[i] = float64(c.controller.Load()) / 127
	}
}
