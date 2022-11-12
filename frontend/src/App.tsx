import {
  useEffect,
  useState,
  useRef,
} from 'react';

import {
  EventsOn,
  EventsOff,
} from '../wailsjs/runtime';

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
import styled from 'styled-components';

import Inspector from './components/Inspector';

const AppContainer = styled.div`
  display: flex;
  flex-direction: column;
`;

const StyledContainer = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: stretch;
`;

function App() {
  const [graphSeqNum, setGraphSeqNum] = useState(0);
  const [graphDot, setGraphDot] = useState<string|undefined>();

  const sampleRate = 44100;

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
    <AppContainer id="App">
      <h2>Synthesizer</h2>
      <label>Output Gain</label>
      <FloatInput onValueChange={handleGainChange} />
      <StyledContainer>
        <div>
          {/* {graphDot && <Graphviz options={{width: 1000}} dot={graphDot} />} */}
          OK
        </div>
        <Inspector signals={[
          {
            id: "output",
            label: "Output",
            sampleRate: sampleRate,
            samplesCallback: (cb) => {
              EventsOn("samples", cb);
              return () => {
                EventsOff("samples");
              };
            },
          },
        ]} />
      </StyledContainer>
    </AppContainer>
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
