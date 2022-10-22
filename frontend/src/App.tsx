import {
  useEffect,
  useState,
} from 'react';
import logo from './assets/images/logo-universal.png';
import './App.css';
import {
  SetGain,
  GraphDot,
  SetShowSpectrum,
  SetShowSpectrumHist,
  SetShowOscilloscope,
  SetOscilloscopeWindow,
  SetOscilloscopeFreq,
} from "../wailsjs/go/main/App";

import Graphviz from 'graphviz-react';

function App() {
  const [graphSeqNum, setGraphSeqNum] = useState(0);
  const [graphDot, setGraphDot] = useState<string|undefined>();

  const [showFFT, setShowFFT] = useState(true);
  const [showFFTHist, setShowFFTHist] = useState(true);

  const [showOscilloscope, setShowOscilloscope] = useState(true);
  const [oscilloscopeWindow, setOscilloscopeWindow] = useState(0.01);
  const [oscilloscopeFreq, setOscilloscopeFreq] = useState(1);

  useEffect(() => {
    const ival = setInterval(() => {
      setGraphSeqNum(graphSeqNum + 1);
    }, 1000);

    const updateGraph = async () => {
      setGraphDot(await GraphDot());
    };
    updateGraph();

    return () => {
      clearInterval(ival);
    };
  }, [graphSeqNum]);

  const handleGainChange = (gain: number) => {
    SetGain(gain);
  };

  return (
    <div id="App">
      <h2>Synthesizer</h2>
      <label>Output Gain</label>
      <FloatInput onValueChange={handleGainChange} />
      <div>
        <label>Show FFT</label>
        <input type="checkbox" checked={showFFT} onChange={(e) => {
          setShowFFT(e.target.checked);
          SetShowSpectrum(e.target.checked);
        }} />
        <br />
        <label>Show FFT Histogram</label>
        <input type="checkbox" checked={showFFTHist} onChange={(e) => {
          setShowFFTHist(e.target.checked);
          SetShowSpectrumHist(e.target.checked);
        }} />
      </div>
      {/* horizontal line */}
      <hr />
      <div>
        <label>Show Oscilloscope</label>
        <input type="checkbox" checked={showOscilloscope} onChange={(e) => {
          setShowOscilloscope(e.target.checked);
          SetShowOscilloscope(e.target.checked);
        }} />
        <br />
        <label>Oscilloscope Window</label>
        <input type="number" step="0.001" min="0.001" max="0.5" value={oscilloscopeWindow} onChange={(e) => {
          setOscilloscopeWindow(parseFloat(e.target.value));
          SetOscilloscopeWindow(parseFloat(e.target.value));
        }} />
        <br />
        <label>Oscilloscope Frequency (Hz)n</label>
        <input type="number" step="0.25" min="0.25" max="100" value={oscilloscopeFreq} onChange={(e) => {
          setOscilloscopeFreq(parseFloat(e.target.value));
          SetOscilloscopeFreq(parseFloat(e.target.value));
        }} />
      </div>
      <div>
        {graphDot && <Graphviz options={{width: 1000}} dot={graphDot} />}
      </div>
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
