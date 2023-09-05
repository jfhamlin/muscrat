package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	inPorts := midi.GetInPorts()
	fmt.Printf("found %d input ports\n", len(inPorts))
	if len(inPorts) == 0 {
		return
	}

	port := inPorts[0]
	fmt.Printf("opening port [%d] %q\n", port.Number(), port)

	stop, err := midi.ListenTo(port, func(msg midi.Message, timestampms int32) {
		switch msg.Type() {
		case midi.NoteOnMsg:
			var channel, key, velocity uint8
			msg.GetNoteOn(&channel, &key, &velocity)
			fmt.Printf("NoteOn: channel=%d key=%d velocity=%d\n", channel, key, velocity)
		case midi.NoteOffMsg:
			var channel, key, velocity uint8
			msg.GetNoteOff(&channel, &key, &velocity)
			fmt.Printf("NoteOff: channel=%d key=%d velocity=%d\n", channel, key, velocity)
		case midi.PitchBendMsg:
			var channel uint8
			var relative int16
			var absolute uint16
			msg.GetPitchBend(&channel, &relative, &absolute)
			fmt.Printf("PitchBend: channel=%d relative=%d absolute=%d\n", channel, relative, absolute)
		case midi.AfterTouchMsg:
			var channel, pressure uint8
			msg.GetAfterTouch(&channel, &pressure)
			fmt.Printf("AfterTouch: channel=%d pressure=%d\n", channel, pressure)
		case midi.ControlChangeMsg:
			var channel, controller, value uint8
			msg.GetControlChange(&channel, &controller, &value)
			fmt.Printf("ControlChange: channel=%d controller=%d value=%d\n", channel, controller, value)
		default:
			fmt.Printf("unknown type %s\n", msg.Type())
		}
	})
	if err != nil {
		panic(err)
	}
	defer stop()

	// wait for signal to stop
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, os.Interrupt, syscall.SIGTERM)
	<-osSignal
}
