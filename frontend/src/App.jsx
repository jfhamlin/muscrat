import {
  GetSampleRate,
} from "../bindings/github.com/jfhamlin/muscrat/muscratservice";

import {
  useState,
  useEffect,
} from 'react';

import {
  BuffersProvider,
  createBuffersStore,
} from "./contexts/buffers";

import Editor from "./components/Editor";
import Toolbar from "./components/Toolbar";
import Sidebar from "./components/Sidebar";
import Docs from "./components/Docs";

function App() {
  return (
    <BuffersProvider createStore={createBuffersStore}>
      <div className="flex flex-row w-screen h-screen overflow-hidden bg-white">
        <Sidebar />
        <div className="flex flex-col flex-1 h-full overflow-hidden">
          <Toolbar />
          <Editor />
          {/* horizontal line */}
          <div className="border-t border-gray-300" />
          {/* <Docs /> */}
        </div>
      </div>
    </BuffersProvider>
  )
}

export default App;
