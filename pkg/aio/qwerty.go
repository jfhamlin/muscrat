package aio

import (
	"context"
	"math"
	"time"

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

func NewQwertyMIDI() ugen.UGen {
	return &qwertyMIDI{}
}

func (q *qwertyMIDI) Start(ctx context.Context) error {
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

func (q *qwertyMIDI) Stop(ctx context.Context) error {
	q.cancel()
	return nil
}

func (q *qwertyMIDI) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := range out {
		out[i] = q.lastNote
	}
}

func (q *qwertyMIDI) AsTrigger() ugen.UGen {
	return &qwertyTrig{lastTrig: &q.lastTrig}
}

func (q *qwertyTrig) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := range out {
		if *q.lastTrig {
			out[i] = 1
		} else {
			out[i] = -1
		}
	}
}

func midiFreq(n float64) float64 {
	return 440 * math.Pow(2, (n-69)/12)
}

var (
	midiCharMap = map[byte]float64{
		'a':  midiFreq(48),
		'w':  midiFreq(49),
		's':  midiFreq(50),
		'e':  midiFreq(51),
		'd':  midiFreq(52),
		'f':  midiFreq(53),
		't':  midiFreq(54),
		'g':  midiFreq(55),
		'y':  midiFreq(56),
		'h':  midiFreq(57),
		'u':  midiFreq(58),
		'j':  midiFreq(59),
		'k':  midiFreq(60),
		'o':  midiFreq(61),
		'l':  midiFreq(62),
		'p':  midiFreq(63),
		';':  midiFreq(64),
		'\'': midiFreq(65),
		']':  midiFreq(66),
	}
)

func mapToFreq(b byte) (float64, bool) {
	if freq, ok := midiCharMap[b]; ok {
		return freq, true
	}
	return 0, false
}
