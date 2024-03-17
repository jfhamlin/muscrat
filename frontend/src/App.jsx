import {
  useState,
  useEffect,
} from 'react';

import { EventsOn } from '../wailsjs/runtime';

import logo from './assets/images/muscrat.svg';

import {
  BuffersProvider,
  createBuffersStore,
} from "./contexts/buffers";

import Editor from "./components/Editor";
import Toolbar from "./components/Toolbar";
import VolumeMeter from "./components/VolumeMeter";

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
  useEffect(() => {
    const audioResources = createAudioResources();
    setAudioResources(audioResources);

    const unsubscribe = EventsOn("samples", (samples) => {
      const samplesChannel0 = Float32Array.from(samples[0]);
      const samplesChannel1 = Float32Array.from(samples[1]);

      const context = audioResources.context;
      const analyser = audioResources.analyser;

      const buffer = context.createBuffer(2, samples.length, 44100);
      buffer.copyToChannel(samplesChannel0, 0);
      buffer.copyToChannel(samplesChannel1, 1);

      const source = context.createBufferSource();
      source.buffer = buffer;
      source.connect(analyser);
      source.start();
    });

    return () => {
      unsubscribe();
      audioResources.context.close();
    };
  }, []);

  return (
    <BuffersProvider createStore={createBuffersStore}>
      <div className="flex flex-row">
        <div className="flex flex-col items-center">
          <img src={logo} className="w-32 max-w-32"/>
          <VolumeMeter analyser={audioResources?.analyser} />
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
