import {
  useEffect,
  useState,
  useRef,
} from 'react';

import {
  EventsOn,
  EventsOff,
} from '../wailsjs/runtime';

import { Line, Bar } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  LogarithmicScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js';

ChartJS.register(
  CategoryScale,
  LinearScale,
  LogarithmicScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
);

// @ts-ignore
import { fft, util as fftUtil } from 'fft-js';

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

  const [samples, setSamples] = useState<[number]>([0]);

  const sampleBuffer = useRef<[number]>([0]);
  const lastUpdate = useRef<number>(0);
  const lastFftUpdate = useRef<number>(0);

  const [freqBins, setFreqBins] = useState<[number]>([0]);
  const [freqBinLabels, setFreqBinLabels] = useState<[number]>([0]);

  const sampleRate = 44100;

  useEffect(() => {
    EventsOn("samples", (samples: any) => {
      sampleBuffer.current = samples;
      const now = Date.now();
      if (showOscilloscope && now - lastUpdate.current > (1000.0 / oscilloscopeFreq)) {
        setSamples(samples);
        lastUpdate.current = now;
      }

      if (showFFT && now - lastFftUpdate.current > (1000.0 / 4)) { // n times per second
        // apply a hann window
        const fftSamps = samples.map(
          (s: number, i: number) => s * (0.5 - 0.5 * Math.cos(2 * Math.PI * i / samples.length))
        );

        const bins = fft(fftSamps);
        setFreqBinLabels(fftUtil.fftFreq(bins, sampleRate).map((f: number) => Math.round(f)));
        setFreqBins(fftUtil.fftMag(bins));

        lastFftUpdate.current = now;
      }
    });
    return () => {
      EventsOff("samples");
    };
  }, [oscilloscopeFreq]);

  useEffect(() => {
    const ival = setInterval(() => {
      setGraphSeqNum(graphSeqNum + 1);
    }, 1000);

    const updateGraph = async () => {
      const dot = await GraphDot();
      if (dot !== graphDot) {
        setGraphDot(dot);
      }
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
        <label>Oscilloscope Frequency (Hz)</label>
        <input type="number" step="0.25" min="0.1" max="100" value={oscilloscopeFreq} onChange={(e) => {
          setOscilloscopeFreq(parseFloat(e.target.value));
          SetOscilloscopeFreq(parseFloat(e.target.value));
        }} />
      </div>
      <div style={{backgroundColor: "white"}}>
        {showOscilloscope ? <LineChart samples={samples} /> : null}
        {showFFT ? <Histogram bins={freqBins} labels={freqBinLabels} /> : null}
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

function LineChart(props: {samples: [number]}) {
  const options = {
    responsive: true,
    plugins: {
      legend: {
        position: 'top' as const,
      },
      title: {
        display: true,
        text: 'Samples',
      },
    },
  };
  const data = {
    labels: props.samples.map((_, i) => i),
    datasets: [{
      label: "Samples",
      data: props.samples,
      borderColor: 'rgb(53, 162, 235)',
      backgroundColor: 'rgba(53, 162, 235, 0.5)',
      pointStyle: 'cross',
      radius: 0,
    }]
  }
  return <Line options={options} data={data} updateMode="none" />;
}

function Histogram(props: {labels: [number]|undefined, bins: [number]}) {
  const options = {
    responsive: true,
    plugins: {
      legend: {
        position: 'top' as const,
      },
      title: {
        display: true,
        text: 'Frequency Bins',
      },
    },
    scales: {
      x: {
        display: true,
        type: 'logarithmic',
      },
    },
  };
  const data = {
    labels: props.labels ?? props.bins.map((_, i) => i),
    datasets: [{
      label: "Frequency Bins",
      data: props.bins,
      borderColor: 'rgb(53, 162, 235)',
      backgroundColor: 'rgba(53, 162, 235, 0.5)',
      pointStyle: 'cross',
      radius: 0,
    }]
  }
  //@ts-ignore
  return <Line options={options} data={data} />;
}

export default App
