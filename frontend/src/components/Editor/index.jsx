import {
  useRef,
} from 'react';

import {
  OpenFileDialog,
  SaveFile,
} from "../../../wailsjs/go/main/App";

import Editor from '@monaco-editor/react';

import {
  useBuffersStore,
} from "../../contexts/buffers";

const DEFAULT_CODE = `(ns user
  (:require [mrat.core :refer :all]
    [mrat.scales :refer :all]
    [mrat.midi :refer :all]))

(play (sin 200))
`;

export default (props) => {
  const buffersStore = useBuffersStore();

  const selectedBufferName = buffersStore.selectedBufferName;
  const selectedBuffer = buffersStore.buffers[selectedBufferName];
  const code = selectedBuffer?.content || DEFAULT_CODE;

  const selectedBufferNameRef = useRef(selectedBufferName);
  selectedBufferNameRef.current = selectedBufferName;

  const editorRef = useRef(null);
  const handleEditorDidMount = (editor, monaco) => {
    editorRef.current = editor;

    // add a key binding for cmd+s
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
      const name = selectedBufferNameRef.current;
      SaveFile(name, editorRef.current.getValue()).then(() => {
        buffersStore.cleanBuffer(name, editorRef.current.getValue());
      }).catch((err) => {
        console.log(err);
      });
    });
  };

  const handleEditorChange = (value, event) => {
    buffersStore.updateBuffer(selectedBufferName, value);
  };

  return (
    <div>
      <Editor height="90vh"
              defaultLanguage="clojure"
              path={selectedBufferName}
              defaultValue={code}
              onChange={handleEditorChange}
              onMount={handleEditorDidMount} />
      <div>{selectedBufferName}{selectedBuffer?.dirty ? "*" : ""}</div>
    </div>
  );
}
