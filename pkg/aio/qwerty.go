package aio

import (
	"context"
	"math"
	"time"

	"github.com/jfhamlin/muscrat/pkg/ugen"
)

type (
	QwertyMIDI struct {
		cancel   context.CancelFunc
		lastNote float64
		lastTrig bool
	}

	qwertyTrig struct {
		lastTrig *bool
	}
)

var (
	StdinChan = make(chan byte, 256)
)

func NewQwertyMIDI() ugen.UGen {
	return &QwertyMIDI{}
}

func (q *QwertyMIDI) Start(ctx context.Context) error {
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
					note, ok := midiCharMap[b]
					if ok {
						q.lastNote = float64(note)
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

func (q *QwertyMIDI) Stop(ctx context.Context) error {
	q.cancel()
	return nil
}

func (q *QwertyMIDI) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := range out {
		out[i] = q.lastNote
	}
}

func (q *QwertyMIDI) AsTrigger() ugen.UGen {
	return &qwertyTrig{lastTrig: &q.lastTrig}
}

func (q *qwertyTrig) Gen(ctx context.Context, cfg ugen.SampleConfig, out []float64) {
	for i := range out {
		if *q.lastTrig {
			out[i] = 1
		} else {
			out[i] = 0
		}
	}
}

func midiFreq(n float64) float64 {
	return 440 * math.Pow(2, (n-69)/12)
}

var (
	midiCharMap = func() map[byte]int {
		const f2 = 41
		keys := []byte{
			/*
			 f         g         a         b    c         d         e*/
			'a', 'w', 's', 'e', 'd', 'r', 'f', 'g', 'y', 'h', 'u', 'j',
			'k', 'o', 'l', 'p', ';', '[', '\''}

		ret := map[byte]int{}
		for i, b := range keys {
			ret[b] = f2 + i
		}
		return ret
	}()
)
