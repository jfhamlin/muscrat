import {useState} from 'react';

import { EventsOn } from '../wailsjs/runtime';

import logo from './assets/images/muscrat.svg';

import {
  BuffersProvider,
  createBuffersStore,
} from "./contexts/buffers";

import Editor from "./components/Editor";
import Toolbar from "./components/Toolbar";
import VolumeMeter from "./components/VolumeMeter";

function App() {
  const subscribeToSampleBuffer = (fn) => {
    // subscribe to "sampels" with wails EventsOn
    const unsubscribe = EventsOn("samples", (samples) => {
      fn(samples);
    });

    return () => {
      unsubscribe();
    };
  }

  return (
    <BuffersProvider createStore={createBuffersStore}>
      <div className="flex flex-row">
        <div className="flex flex-col items-center">
          <img src={logo} className="w-32 max-w-32"/>
          <VolumeMeter subscribeToSampleBuffer={subscribeToSampleBuffer} />
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
