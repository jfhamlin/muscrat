package aio

import (
	"context"
	"time"

	"github.com/jfhamlin/muscrat/pkg/midi"
	"github.com/jfhamlin/muscrat/pkg/ugen"
)

var (
	StdinChan = make(chan byte, 256)
)

type qwertyMIDI struct {
	cancel   context.CancelFunc
	lastNote float64
	lastTrig bool
}

type qwertyTrig struct {
	lastTrig *bool
}

func NewQwertyMIDI() ugen.SampleGenerator {
	return &qwertyMIDI{}
}

func (q *qwertyMIDI) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	q.cancel = cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				select {
				case b := <-StdinChan:
					note, ok := mapToFreq(b)
					if ok {
						q.lastNote = note
						q.lastTrig = true
					} else {
						q.lastTrig = false
					}
				case <-time.After(100 * time.Millisecond):
					// continue if we're not getting keys
				}
			}
		}
	}()
	return nil
}

func (q *qwertyMIDI) Stop() error {
	q.cancel()
	return nil
}

func (q *qwertyMIDI) GenerateSamples(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		res[i] = q.lastNote
	}
	return res
}

func (q *qwertyMIDI) AsTrigger() ugen.SampleGenerator {
	return &qwertyTrig{lastTrig: &q.lastTrig}
}

func (q *qwertyTrig) GenerateSamples(ctx context.Context, cfg ugen.SampleConfig, n int) []float64 {
	res := make([]float64, n)
	for i := range res {
		if *q.lastTrig {
			res[i] = 1
		}
	}
	return res
}

var (
	midiCharMap = map[byte]float64{
		'a':  midi.C3.Frequency,
		'w':  midi.Cs3.Frequency,
		's':  midi.D3.Frequency,
		'e':  midi.Ds3.Frequency,
		'd':  midi.E3.Frequency,
		'f':  midi.F3.Frequency,
		't':  midi.Fs3.Frequency,
		'g':  midi.G3.Frequency,
		'y':  midi.Gs3.Frequency,
		'h':  midi.A3.Frequency,
		'u':  midi.As3.Frequency,
		'j':  midi.B3.Frequency,
		'k':  midi.C4.Frequency,
		'o':  midi.Cs4.Frequency,
		'l':  midi.D4.Frequency,
		'p':  midi.Ds4.Frequency,
		';':  midi.E4.Frequency,
		'\'': midi.F4.Frequency,
		']':  midi.Fs4.Frequency,
	}
)

func mapToFreq(b byte) (float64, bool) {
	if freq, ok := midiCharMap[b]; ok {
		return freq, true
	}
	return 0, false
}
