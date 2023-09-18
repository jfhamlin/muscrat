import {
  useEffect,
  useState,
  useRef,
} from 'react';

import { Line, Bar } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  Filler,
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
  Filler,
  LinearScale,
  LogarithmicScale,
  PointElement,
  LineElement,
  BarElement,
  Title,
  Tooltip,
  Legend,
);

import styled from 'styled-components';

import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Slider from '@mui/material/Slider';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

import VolumeSlider from '../VolumeSlider';

// @ts-ignore
import { fft, util as fftUtil } from 'fft-js';

interface SignalInfo {
  id: string;
  label: string;
  samplesCallback: (cb: (samples: number[]) => void) => () => void;
  sampleRate: number;
}

interface InspectorProps {
  signals: SignalInfo[];
  volume: number;
  setVolume: (volume: number) => void;
}

const StyledContainer = styled.div`
  color: black;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: #fff;
  padding: 1rem;


  width: 100%;
`;

function ChartBox(props: any) {
  return (
    <Box
      sx={{
        border: 1,
        borderColor: 'divider',
        marginBottom: '0.25rem',
        padding: '1rem',
        width: '512px',
      }}
      {...props}
    >
      {props.children}
    </Box>
  );
}

export default function Inspector(props: InspectorProps) {
  return (
    <StyledContainer>
      <VolumeSlider volume={props.volume} onChange={props.setVolume} />
      {props.signals.map((signal) => (
        <SignalInspector key={signal.id} signal={signal} />
      ))}
    </StyledContainer>
  );
}

function SignalInspector(props: { signal: SignalInfo }) {
  const { signal } = props;

  // Visualization options

  const [oscilloscopeWindow, setOscilloscopeWindow] = useState(0.01);
  const handleOscilloscopeWindowChange = (
    event: Event,
    newValue: number | number[],
  ) => {
    setOscilloscopeWindow(newValue as number);
  };

  const [oscilloscopeFreq, setOscilloscopeFreq] = useState(10);
  const handleOscilloscopeFreqChange = (
    event: Event,
    newValue: number | Array<number>,
  ) => {
    setOscilloscopeFreq(newValue as number);
  };

  const [fftFreq, setFftFreq] = useState(15);
  const handleFftFreqChange = (
    event: Event,
    newValue: number | Array<number>,
  ) => {
    setFftFreq(newValue as number);
  };

  // Visualization data
  const [samples, setSamples] = useState<number[]>([]);
  const [freqBins, setFreqBins] = useState<[number]>([0]);
  const [freqBinLabels, setFreqBinLabels] = useState<[number]>([0]);

  const lastUpdate = useRef<number>(0);
  const lastFftUpdate = useRef<number>(0);

  const numOscSamples = signal.sampleRate * oscilloscopeWindow;
  const oscSamples = useRef<number[]>([]);
  const fftSamples = useRef<number[]>([]);

  const samplesCallback = signal.samplesCallback;
  useEffect(() => {
    const MAX_FFT_SAMPLES = 4096;
    return samplesCallback((newSamples) => {
      const now = Date.now();

      oscSamples.current = oscSamples.current.concat(newSamples);
      if (oscSamples.current.length > numOscSamples) {
        oscSamples.current = oscSamples.current.slice(-numOscSamples);
      }
      if (now - lastUpdate.current > (1000.0 / oscilloscopeFreq)) {
        setSamples(oscSamples.current);
        lastUpdate.current = now;
      }

      fftSamples.current = fftSamples.current.concat(newSamples);
      if (fftSamples.current.length > MAX_FFT_SAMPLES) {
        fftSamples.current = fftSamples.current.slice(-MAX_FFT_SAMPLES);
      }
      if (now - lastFftUpdate.current > (1000.0 / fftFreq)) {
        // apply a hann window
        const fftSamps = fftSamples.current.map(
          (s: number, i: number) => s * (0.5 - 0.5 * Math.cos(2 * Math.PI * i / fftSamples.current.length))
        );

        const bins = fft(fftSamps);
        setFreqBinLabels(fftUtil.fftFreq(bins, signal.sampleRate).map((f: number) => Math.round(f)));
        const mags = fftUtil.fftMag(bins);
        setFreqBins(mags);

        lastFftUpdate.current = now;
      }
    });
  }, [samplesCallback, numOscSamples, oscilloscopeFreq, fftFreq, signal.sampleRate]);

  return (
    <>
    <h3>{props.signal.label}</h3>
       <ChartBox>
         <LabeledSlider
           label="Window Width (s)"
           value={oscilloscopeWindow}
           onChange={handleOscilloscopeWindowChange}
           min={0.001}
           max={0.5}
           step={0.001} />
         <LabeledSlider
           label="Update Frequency (Hz)"
           value={oscilloscopeFreq}
           onChange={handleOscilloscopeFreqChange}
           min={0.5}
           max={10}
           step={0.5} />
         <LineChart samples={samples} />
       </ChartBox>

       <ChartBox>
         <LabeledSlider
           label="Update Frequency (Hz)"
           value={fftFreq}
           onChange={handleFftFreqChange}
           min={1}
           max={15}
           step={1} />
         <Histogram bins={freqBins} labels={freqBinLabels} />
       </ChartBox>
    </>
  );
}

function LabeledSlider(props: {
  label: string;
  min: number;
  max: number;
  step: number;
  value: number;
  onChange: (event: Event, newValue: number | Array<number>) => void;
}) {
  return (
    <Box>
      <Typography gutterBottom>{props.label}</Typography>
      <Slider
        aria-label={props.label}
        size="small"
        min={props.min}
        max={props.max}
        step={props.step}
        value={props.value}
        onChange={props.onChange}
        valueLabelDisplay="auto" />
    </Box>
  );
}

function LineChart(props: {samples: number[]}) {
  const options = {
    responsive: true,
    plugins: {
      legend: {
        display: false,
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
      borderWidth: 1,
      borderColor: 'rgb(53, 162, 235)',
      backgroundColor: 'rgba(53, 162, 235, 0.5)',
      pointStyle: 'cross',
      radius: 0,
    }]
  }
  return <Line options={options} data={data} updateMode="none" />;
}

function Histogram(props: {labels: number[] | undefined, bins: number[]}) {
  const options = {
    animation: false,
    responsive: true,
    plugins: {
      legend: {
        display: false,
      },
      title: {
        display: true,
        text: 'Frequency Spectrum',
      },
    },
    scales: {
      x: {
        display: true,
        type: 'logarithmic',
      },
      y: {
        display: true,
        min: 0,
        //type: 'logarithmic',
      },
    },
  };
  const data = {
    labels: props.labels ?? props.bins.map((_, i) => i),
    datasets: [{
      label: "Frequency Spectrum",
      data: props.bins,
      borderColor: 'rgb(53, 162, 235)',
      backgroundColor: 'rgba(53, 162, 235, 0.5)',
      pointStyle: 'cross',
      borderWidth: 1,
      radius: 0,
      cubicInterpolationMode: 'monotone',
      fill: true,
    }]
  }
  //@ts-ignore
  return <Line options={options} data={data} />;
}
