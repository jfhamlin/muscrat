import {
  useEffect,
  useState,
  useRef,
} from 'react';

import {
  EventsOn,
  EventsOff,
} from '../wailsjs/runtime';

import logo from './assets/images/muscrat.svg';

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
// import UGenGraph from './components/UGenGraph';

const AppContainer = styled.div`
  display: flex;
  flex-direction: column;
  background-color: #fff;
`;

const StyledContainer = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: stretch;
`;

const graph = {
  nodes: [
    { id: 1, label: "Node 1", title: "node 1 tootip text" },
    { id: 2, label: "Node 2", title: "node 2 tootip text" },
    { id: 3, label: "Node 3", title: "node 3 tootip text" },
    { id: 4, label: "Node 4", title: "node 4 tootip text" },
    { id: 5, label: "Node 5", title: "node 5 tootip text" }
  ],
  edges: [
    { from: 1, to: 2 },
    { from: 1, to: 3 },
    { from: 2, to: 4 },
    { from: 2, to: 5 }
  ]
};

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

  const [gain, setGain] = useState(0.5);

  const handleGainChange = (gain: number) => {
    setGain(gain);
    SetGain(gain);
  };

  return (
    <AppContainer id="App">
      <img src={logo} className="App-logo" alt="logo" style={{
        maxHeight: '100px',
        objectFit: 'contain',
      }} />
      <StyledContainer>
        <div>
          {/* <UGenGraph graph={graph} /> */}
        </div>
        <Inspector
          volume={gain}
          setVolume={handleGainChange}
          signals={[
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
