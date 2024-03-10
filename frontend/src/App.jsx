import {useState} from 'react';
import './App.css';

import {
  BuffersProvider,
  createBuffersStore,
} from "./contexts/buffers";

import Editor from "./components/Editor";

function App() {
    return (
      <BuffersProvider createStore={createBuffersStore}>
        <div id="App">
          <Editor />
        </div>
      </BuffersProvider>
    )
}

export default App;
