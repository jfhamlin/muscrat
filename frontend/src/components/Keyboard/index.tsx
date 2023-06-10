import React from 'react';

// @ts-ignore
import { Piano, KeyboardShortcuts, MidiNumbers } from 'react-piano';

import 'react-piano/dist/styles.css';

const noteRange = {
  first: MidiNumbers.fromNote('c3'),
  last: MidiNumbers.fromNote('f4'),
};
const keyboardShortcuts = KeyboardShortcuts.create({
  firstNote: noteRange.first,
  lastNote: noteRange.last,
  keyboardConfig: KeyboardShortcuts.HOME_ROW,
});

interface KeyboardProps {
  onEvent: (evt: any) => void,
};

const Keyboard = (props: KeyboardProps) => {
  return (
    <Piano
      noteRange={noteRange}
      width={500}
      playNote={(midiNumber: any) => {
        props.onEvent({ type: 'noteOn', midiNumber });
      }}
      stopNote={(midiNumber: any) => {
        props.onEvent({ type: 'noteOff', midiNumber });
      }}
      keyboardShortcuts={keyboardShortcuts}
    />
  );
};

export default Keyboard;
