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
import Sidebar from "./components/Sidebar";
import Docs from "./components/Docs";
import Debugger from "./components/Debugger";
import Knobs from "./components/Knobs";
import Toolbar from "./components/Toolbar";

function App() {
  return (
    <BuffersProvider createStore={createBuffersStore}>
      <div className="flex flex-col w-screen h-screen overflow-hidden bg-background-primary select-none cursor-default">
        <div className="mt-2">
          <Toolbar />
        </div>
        <div className="flex flex-col flex-1 h-full overflow-hidden mt-1">
          <Splitter gutterClassName="bg-transparent border-t border-l border-gray-300/25"
                    initialSizes={[67, 33]}>
            <Splitter direction={SplitDirection.Vertical}
                      gutterClassName="bg-transparent border-t border-gray-300/25"
                      initialSizes={[67, 33]}>
              <div className="h-full">
                <Editor />
              </div>
              <Debugger />
            </Splitter>
            <div className="h-full border-t border-gray-300/25">
              <Sidebar />
            </div>
          </Splitter>
        </div>
      </div>
    </BuffersProvider>
  )
}

export default App;
