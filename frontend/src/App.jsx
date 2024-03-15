import {useState} from 'react';
import './App.css';

import {
  BuffersProvider,
  createBuffersStore,
} from "./contexts/buffers";

import Editor from "./components/Editor";
import Toolbar from "./components/Toolbar";

function App() {
    return (
      <BuffersProvider createStore={createBuffersStore}>
        <div id="App">
          <Toolbar />
          <Editor />
        </div>
      </BuffersProvider>
    )
}

export default App;
