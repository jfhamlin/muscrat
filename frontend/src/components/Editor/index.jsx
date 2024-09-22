import {
  useRef,
  useState,
  useEffect,
} from 'react';

import {
  OpenFileDialog,
  SaveFile,
  GetNSPublics,
} from "../../../bindings/github.com/jfhamlin/muscrat/muscratservice";

import Editor from '@monaco-editor/react';

import {
  useBuffersStore,
} from "../../contexts/buffers";

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

  // resize the editor when the window is resized
  useEffect(() => {
    if (!editor) {
      return;
    }

    // strategy from
    // https://berezuzu.medium.com/resizable-monaco-editor-3e922ad54e4
    const handleResize = () => {
      editor.layout({ width: 0, height: 0 });
    };

    window.addEventListener('resize', handleResize);
    return () => {
      window.removeEventListener('resize', handleResize);
    };
  }, [editor]);

  const options = {
    minimap: {
      enabled: false,
    },
  };

  // monaco editor layout is a pain to manage
  return (
    <>
      <div>
        <Editor options={options}
                width="100%"
                height="90vh"
                defaultLanguage="clojure"
                path={selectedBufferName}
                defaultValue={code}
                onChange={handleEditorChange}
                onMount={handleEditorDidMount} />
      </div>
      <div className="text-xs text-gray-500 border-t border-gray-300">
        {selectedBufferName}{selectedBuffer?.dirty ? "*" : ""}
      </div>
    </>
  );
}
