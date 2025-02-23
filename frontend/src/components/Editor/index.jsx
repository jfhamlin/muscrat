import {
  useRef,
  useState,
  useEffect,
  useCallback,
} from 'react';

import {
  OpenFileDialog,
  SaveFile,
  GetNSPublics,
} from "../../../bindings/github.com/jfhamlin/muscrat/muscratservice";

import Editor, { loader } from '@monaco-editor/react';

import {
  useBuffersStore,
} from "../../contexts/buffers";

import { tailwindTheme } from "../../theme";

const theme = {
  base: 'vs-dark',
  inherit: true,
  rules: [],
  colors: {
    'editor.background': tailwindTheme.background.secondary,
  },
}

loader.init().then((monaco) => {
  monaco.editor.defineTheme('muscrat', theme);
});

export default (props) => {
  const buffersStore = useBuffersStore();

  const selectedBufferName = buffersStore.selectedBufferName;
  const selectedBuffer = buffersStore.buffers[selectedBufferName];
  const code = selectedBuffer?.content;

  const selectedBufferNameRef = useRef(selectedBufferName);
  selectedBufferNameRef.current = selectedBufferName;

  const [editor, setEditor] = useState(null);
  const handleEditorDidMount = (editor, monaco) => {
    setEditor(editor);

    // add a key binding for cmd+s
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      const name = selectedBufferNameRef.current;
      const content = editor.getValue();
      SaveFile(name, content).then((fileName) => {
        buffersStore.updateBuffer(fileName, content, false);
        buffersStore.selectBuffer(fileName);
      }).catch((err) => {
        console.log(err);
      });
    });
  };

  const handleEditorChange = (value, event) => {
    buffersStore.updateBuffer(selectedBufferName, value);
  };

  const editorContainerRef = useCallback((container) => {
    if (!container) return;

    console.log("editorContainerRef", container);

    const resizeObserver = new ResizeObserver(() => {
      console.log("resize");
      if (!container || !editor) {
        return;
      }
      const { width, height } = container.getBoundingClientRect();
      if (width === 0 || height === 0) {
        return;
      }
      // set the editor to the size of the container
      editor.layout({ width, height });
    });
    resizeObserver.observe(container);
  }, []);

  const options = {
    padding: {
      top: 10,
    },
    folding: false,
    lineNumbersMinChars: 3,
    minimap: {
      enabled: false,
    },
  };

  // monaco editor layout is a pain to manage
  return (
    <>
      <div className="flex flex-col h-full">
        <div className="flex">
          <div className="text-xs text-gray-500 border-t border-r border-gray-300/25 flex-shrink min-w-[2rem] bg-background-secondary px-2 pb-1 pt-1 rounded-t">
            {selectedBufferName ?? '<new>'}{selectedBuffer?.dirty ? "*" : ""}
          </div>
          <div className="border-b border-gray-300/25 flex-grow" />
        </div>
        <div className="bg-background-secondary flex-grow overflow-hidden" ref={editorContainerRef}>
          <Editor options={options}
                  theme={'muscrat'}
                  defaultLanguage="clojure"
                  path={selectedBufferName}
                  defaultValue={code}
                  onChange={handleEditorChange}
                  onMount={handleEditorDidMount} />
        </div>
      </div>
    </>
  );
}
