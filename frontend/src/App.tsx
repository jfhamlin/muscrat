import {
  useEffect,
  useState,
} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {
  SetGain,
  GetNotes,
  RegisterWaveformCallback,
} from "../wailsjs/go/main/App";

function App() {
  return (
    <div id="App">
      <h2>Synthesizer</h2>
    </div>
  )
}

export default App
