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

import Editor from "./components/Editor";
import Toolbar from "./components/Toolbar";
import VolumeMeter from "./components/VolumeMeter";
import Spectrogram from "./components/Spectrogram";
import Oscilloscope from "./components/Oscilloscope";

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
      <div className="flex flex-row">
        <div className="flex flex-col items-center w-60">
          <img src={logo} className="w-32 max-w-32"/>
          <div className="flex flex-row">
            <VolumeMeter analyser={audioResources?.analyser} />
            <div>
              <div className="h-40 w-60">
                <Oscilloscope analyser={audioResources?.analyser} />
              </div>
              <div className="h-60 w-full">
                <Spectrogram analyser={audioResources?.analyser} sampleRate={sampleRate} />
              </div>
            </div>
          </div>
        </div>
        <div className="flex flex-col w-full">
          <Toolbar />
          <Editor />
        </div>
      </div>
    </BuffersProvider>
  )
}

export default App;
