import {
  useState,
  useEffect,
} from 'react';

import Splitter, { SplitDirection } from '@devbookhq/splitter'

import {
  GetSampleRate,
} from "../bindings/github.com/jfhamlin/muscrat/muscratservice";

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
      <div className="flex flex-row w-screen h-screen overflow-hidden bg-background-primary select-none cursor-default">
        <div className="flex flex-col flex-1 h-full overflow-hidden">
          <Toolbar />
          <Splitter direction={SplitDirection.Vertical}>
            <div className="h-full">
              <Editor />
            </div>
            <div className="bg-red-500">whatever forever</div>
          </Splitter>
          {/* <Docs /> */}
        </div>
      </div>
    </BuffersProvider>
  )
}

export default App;
