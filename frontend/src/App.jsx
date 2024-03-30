import {
  useState,
  useEffect,
} from 'react';

import { EventsOn } from '../wailsjs/runtime';

import {
  GetSampleRate,
} from "../wailsjs/go/main/App";

import logo from './assets/images/muscrat.svg';

import {
  BuffersProvider,
  createBuffersStore,
} from "./contexts/buffers";

import Console from "./components/Console";
import Editor from "./components/Editor";
import Toolbar from "./components/Toolbar";
import VolumeMeter from "./components/VolumeMeter";
import Spectrogram from "./components/Spectrogram";
import Oscilloscope from "./components/Oscilloscope";
import HydraView from "./components/HydraView";

const createAudioResources = () => {
  const audioContext = new AudioContext();
  const analyser = audioContext.createAnalyser();

  return {
    context: audioContext,
    analyser,
  };
};

function App() {
  const [audioResources, setAudioResources] = useState(null);
  const [sampleRate, setSampleRate] = useState(44100);

  useEffect(() => {
    GetSampleRate().then((sampleRate) => {
      setSampleRate(sampleRate);
    });
  }, []);

  useEffect(() => {
    const audioResources = createAudioResources();
    setAudioResources(audioResources);

    let nextBufferTime = audioResources.context.currentTime;

    const unsubscribe = EventsOn("samples", (samples) => {
      const samplesChannel0 = Float32Array.from(samples[0]);
      const samplesChannel1 = Float32Array.from(samples[1]);

      const bufferLength = samplesChannel0.length;

      const context = audioResources.context;
      const analyser = audioResources.analyser;

      const buffer = context.createBuffer(2, bufferLength, sampleRate);
      buffer.copyToChannel(samplesChannel0, 0);
      buffer.copyToChannel(samplesChannel1, 1);

      const source = context.createBufferSource();
      source.buffer = buffer;
      source.connect(analyser);
      source.start(nextBufferTime);

      nextBufferTime += bufferLength / sampleRate;
    });

    return () => {
      unsubscribe();
      audioResources.context.close();
    };
  }, [sampleRate]);

  return (
    <BuffersProvider createStore={createBuffersStore}>
      <div className="flex flex-row w-screen h-screen overflow-hidden">
        <div className="flex flex-col items-center">
          <img src={logo} className="w-32 max-w-32 my-4" alt="logo" />
          <div className="flex flex-col w-96 bg-gray-400 h-full">
            {/* <VolumeMeter analyser={audioResources?.analyser} /> */}
            <div className="m-1">
              <div className="h-40 mb-1">
                <Oscilloscope analyser={audioResources?.analyser} />
              </div>
              <div className="h-60 mb-1">
                <Spectrogram analyser={audioResources?.analyser} sampleRate={sampleRate} />
              </div>
              <div className="mb-1 h-full">
                <Console />
              </div>
            </div>
          </div>
        </div>
        <div className="flex flex-col flex-1">
          {/* <HydraView /> */}
          <Toolbar />
          <Editor />
        </div>
      </div>
    </BuffersProvider>
  )
}

export default App;
