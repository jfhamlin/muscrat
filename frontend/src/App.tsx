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
  GraphJSON,
  SetShowSpectrum,
  SetShowSpectrumHist,
  SetShowOscilloscope,
  SetOscilloscopeWindow,
  SetOscilloscopeFreq,
} from "../wailsjs/go/main/App";

import styled from 'styled-components';

import Inspector from './components/Inspector';
import UGenGraph from './components/UGenGraph';
import type { Graph } from './components/UGenGraph';

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
  justify-items: stretch;
`;

const StyledGraph = styled.div`
  width: 100%;
  height: 1000;
`;

const graph = {
  nodes: [
    { id: '1', position: { x: 0, y: 0 }, data: { label: '1' }},
    { id: '2', position: { x: 0, y: 100 }, data: { label: '2' }},
  ],
  edges: [
    { id: 'e1-2', source: '1', target: '2' },
  ]
};

function ugenGraphJsonToGraph(json: string): any {
  const graph = JSON.parse(json);

  const result: Graph = {
    nodes: [],
    edges: [],
  };

  (graph.nodes ?? []).forEach((node: any) => {
    result.nodes.push({
      id: String(node.id),
      position: {
        x: 0,
        y: 0,
      },
      data: {
        label: `[${node.type}] ${node.label}`,
      },
    });
  });
  (graph.edges ?? []).forEach((edge: any) => {
    result.edges.push({
      id: `e${edge.from}-${edge.to}-${edge.toPort}`,
      source: String(edge.from),
      target: String(edge.to),
    });
  });

  return result;
}

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

  const [graphUpdateSeqNum, setGraphUpdateSeqNum] = useState(0);
  const [graphJSON, setGraphJSON] = useState<string>("{}");
  const [graph, setGraph] = useState<any>({ edges: [], nodes: [] });

  useEffect(() => {
    const updateGraph = async () => {
      const json = await GraphJSON();
      if (json !== graphJSON) {
        setGraphJSON(json);
        setGraph(ugenGraphJsonToGraph(json));
      }
      setTimeout(() => setGraphUpdateSeqNum((n) => n + 1), 1000);
    };
    updateGraph();
  }, [graphUpdateSeqNum]);

  return (
    <AppContainer id="App">
      <img src={logo} className="App-logo" alt="logo" style={{
        maxHeight: '100px',
        objectFit: 'contain',
      }} />
      <StyledContainer>
        <StyledGraph>
          <UGenGraph graph={graph} />
        </StyledGraph>
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
