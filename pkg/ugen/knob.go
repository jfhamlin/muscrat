package ugen

import (
	"context"
	"math"
	"sync"
	"sync/atomic"

	"github.com/jfhamlin/muscrat/pkg/pubsub"
)

type (
	// Knob is a ugen representing a control that can be turned to
	// adjust a value.
	Knob struct {
		Name string  `json:"name"`
		ID   uint64  `json:"id"`
		Min  float64 `json:"min"`
		Max  float64 `json:"max"`
		Def  float64 `json:"def"`
		Step float64 `json:"step"`

		valueBits atomic.Uint64

		unsubscribe func()
	}

	// KnobUpdate is a message sent to update a knob.
	KnobUpdate struct {
		ID    uint64  `json:"id"`
		Value float64 `json:"value"`
	}
)

const (
	// KnobValueChangeEvent is the event that is sent when a knob's
	// value changes.
	KnobValueChangeEvent = "knob-value-change"

	// KnobsChangedEvent is the event that is sent when the list of
	// knobs changes.
	KnobsChangedEvent = "knobs-changed"
)

var (
	knobLock   sync.Mutex
	nextKnobID uint64
	knobs      = map[uint64]*Knob{}
)

func GetKnobs() []*Knob {
	knobLock.Lock()
	defer knobLock.Unlock()

	knobsList := make([]*Knob, 0, len(knobs))
	for _, knob := range knobs {
		knobsList = append(knobsList, knob)
	}
	return knobsList
}

// NewKnob returns a new Knob ugen.
func NewKnob(name string, def, min, max, step float64) *Knob {
	knobLock.Lock()
	defer knobLock.Unlock()

	k := &Knob{
		Name: name,
		ID:   nextKnobID,
		Min:  min,
		Max:  max,
		Def:  def,
		Step: step,
	}
	k.valueBits.Store(math.Float64bits(def))

	nextKnobID++

	return k
}

func (k *Knob) Start(ctx context.Context) error {
	knobLock.Lock()
	defer knobLock.Unlock()

	knobs[k.ID] = k

	k.unsubscribe = pubsub.Subscribe(KnobValueChangeEvent, func(event string, data any) {
		update := data.(KnobUpdate)
		if update.ID != k.ID {
			return
		}

		bits := math.Float64bits(update.Value)
		k.valueBits.Store(bits)
	})

	pubsub.Publish(KnobsChangedEvent, nil)

	return nil
}

func (k *Knob) Stop(ctx context.Context) error {
	knobLock.Lock()
	defer knobLock.Unlock()

	delete(knobs, k.ID)

	pubsub.Publish(KnobsChangedEvent, nil)

	if k.unsubscribe != nil {
		k.unsubscribe()
	}
	return nil
}

func (k *Knob) Gen(ctx context.Context, cfg SampleConfig, out []float64) {
	for i := range out {
		bits := k.valueBits.Load()
		value := math.Float64frombits(bits)
		out[i] = value
	}
}
