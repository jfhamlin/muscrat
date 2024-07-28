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

  const [completionDisposable, setCompletionDisposable] = useState(null);
  useEffect(() => {
    return () => {
      if (completionDisposable) {
        completionDisposable.dispose();
      }
    };
  }, [completionDisposable]);

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

    /* GetNSPublics().then((nsPublics) => {
     *   console.log("got publics", nsPublics);
     * }); */

    // custom autocomplete
    const completionItemProvider = monaco.languages.registerCompletionItemProvider('clojure', {
      provideCompletionItems: (model, position, context, token) => {
        return {
          suggestions: [
            {
              label: 'defn',
              kind: monaco.languages.CompletionItemKind.Function,
              insertText: 'defn ${1:name} [${2:args}]\n  ${3:body}',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Define a function',
            },
            {
              label: 'def',
              kind: monaco.languages.CompletionItemKind.Variable,
              insertText: 'def ${1:name} ${2:value}',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Define a variable',
            },
            {
              label: 'let',
              kind: monaco.languages.CompletionItemKind.Variable,
              insertText: 'let [${1:bindings}]\n  ${2:body}',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Define a local variable',
            },
            {
              label: 'if',
              kind: monaco.languages.CompletionItemKind.Keyword,
              insertText: 'if ${1:condition}\n  ${2:then}\n  ${3:else}',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Conditional',
            },
            {
              label: 'when',
              kind: monaco.languages.CompletionItemKind.Keyword,
              insertText: 'when ${1:condition}\n  ${2:body}',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Conditional',
            },
            {
              label: '->',
              kind: monaco.languages.CompletionItemKind.Operator,
              insertText: '-> ${1:value} ${2:fn1} ${3:fn2}',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Thread first',
            },
            {
              label: '->>',
              kind: monaco.languages.CompletionItemKind.Operator,
              insertText: '->> ${1:value} ${2:fn1} ${3:fn2}',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Thread last',
            },
          ],
        };
      },
    });
    setCompletionDisposable(completionItemProvider);
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
      <div className="h-full">
        <Editor options={options}
                height="50vh"
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
