import {
  useEffect,
  useState,
} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {
  SetGain,
  GetNotes,
  SetChord,
  RegisterWaveformCallback,
} from "../wailsjs/go/main/App";

function App() {
  const defaultNotes = ["C4", "E4", "Bb4", "D5"];

  const [noteOptions, setNoteOptions] = useState<string[]>(defaultNotes);
  useEffect(() => {
    GetNotes().then((noteOptions) => {
      setNoteOptions(noteOptions);
    });
  }, []);

  const handleGainChange = (note: number) => {
    SetGain(note);
  };

  const [notes, setNotes] = useState<string[]>(defaultNotes);
  const [noteWeights, setNoteWeights] = useState<number[]>(defaultNotes.map(() => 1));

  const handleNoteChange = (note: string, index: number) => {
    const newNotes = [...notes];
    newNotes[index] = note;
    setNotes(newNotes);
    SetChord(newNotes, noteWeights);
  };

  const handleNoteWeightChange = (weight: number, index: number) => {
    const newNoteWeights = [...noteWeights];
    newNoteWeights[index] = weight;
    setNoteWeights(newNoteWeights);
    SetChord(notes, newNoteWeights);
  };

  return (
    <div id="App">
      <h2>Synthesizer</h2>
      <FloatInput onValueChange={handleGainChange} />
      {
        notes.map((note, index) => (
          <div>
            <DropdownInput
              key={index}
              options={noteOptions}
              defaultValue={note}
              onValueChange={(value) => {
                handleNoteChange(value, index);
              }} />
            <FloatInput onValueChange={(value) => handleNoteWeightChange(value, index) } />
          </div>
        ))
      }
    </div>
  )
}

function FloatInput(props: {onValueChange: (value: number) => void}) {
  const [value, setValue] = useState(0.5);

  const handleValueChange = (event: any) => {
    setValue(Number(event.target.value))
  };

  useEffect(() => {
    props.onValueChange(value);
  }, [value]);

  return (
    <div className="input-box">
      <input type="number" value={value} step="0.1" onChange={handleValueChange} />
    </div>
  );
}

/* React component that shows a dropdown input list of string options
 * (from a prop). A user can select an option, and a callback is called.
 */
function DropdownInput(props: {options: string[], defaultValue: string, onValueChange: (value: string) => void}) {
  const [value, setValue] = useState(props.defaultValue);

  const handleValueChange = (event: any) => {
    setValue(event.target.value);
  };

  useEffect(() => {
    props.onValueChange(value);
  }, [value]);

  return (
    <div className="input-box">
      <select value={value} onChange={handleValueChange}>
        {props.options.map((option) => (
          <option key={option} value={option}>{option}</option>
        ))}
      </select>
    </div>
  );
}


export default App
