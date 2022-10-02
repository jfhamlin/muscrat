import {
  useEffect,
  useState,
} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {
  SetGain,
  RegisterWaveformCallback,
} from "../wailsjs/go/main/App";

function App() {
  const handleGainChange = (gain) => {
    SetGain(gain);
  };

  return (
    <div id="App">
      <h2>Synthesizer</h2>
      <FloatInput onValueChange={handleGainChange} />
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
      <input type="number" value={value} step="0.02" onChange={handleValueChange} />
    </div>
  );
}

export default App
