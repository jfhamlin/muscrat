import {
  OpenFileDialog,
  PlayFile,
  SaveFile,
  Silence,
  ToggleHydraWindow,
} from "../../../bindings/github.com/jfhamlin/muscrat/muscratservice";

import {
  useBuffersStore,
} from "../../contexts/buffers";

const Button = (props) => {
  const onClick = props.onClick;
  let className = "h-10 bg-gray-500 hover:bg-gray-700 text-gray-200 font-bold py-2 px-4 rounded m-1"
  if (props.disabled) {
    className += " bg-gray-500 cursor-not-allowed";
  }
  return (
    <button className={className}
      {...props}>{props.children}</button>
  );
};

export default (props) => {
  const buffersStore = useBuffersStore();

  const selectedBufferName = buffersStore.selectedBufferName;
  const selectedBuffer = buffersStore.buffers[selectedBufferName];

  const handleLoadClick = () => {
    OpenFileDialog().then((response) => {
      const buffer = {
        fileName: response.FileName,
        content: response.Content,
      };
      buffersStore.addBuffer(buffer);
    }).catch((err) => {
      console.log(err);
    });
  };

  const handleSaveClick = () => {
    const name = selectedBufferName;
    const content = selectedBuffer.content;
    SaveFile(name, content).then((fileName) => {
      buffersStore.updateBuffer(fileName, content, false);
      buffersStore.selectBuffer(fileName);
    }).catch((err) => {
      console.log(err);
    });
  };

  const handlePlayClick = () => {
    const selectedBufferName = buffersStore.selectedBufferName;
    const buffer = buffersStore.buffers[selectedBufferName];
    PlayFile(buffer.fileName);
  };

  const handleStopClick = () => {
    Silence();
  };

  const handleNewClick = () => {
    const DEFAULT_CONTENT = `(ns user
  (:use [mrat.core]))`;
    buffersStore.updateBuffer(null, DEFAULT_CONTENT, true);
    buffersStore.selectBuffer(null);
  };

  return (
    <div className="flex flex-row ml-[80px]">
      <Button disabled={!buffersStore.selectedBufferName}
              onClick={handlePlayClick}
              title="Play file">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
          <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" />
        </svg>
      </Button>
      <Button onClick={handleStopClick}
              title="Stop playing">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
          <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 7.5A2.25 2.25 0 0 1 7.5 5.25h9a2.25 2.25 0 0 1 2.25 2.25v9a2.25 2.25 0 0 1-2.25 2.25h-9a2.25 2.25 0 0 1-2.25-2.25v-9Z" />
        </svg>
      </Button>
      <Button onClick={ToggleHydraWindow}
              tytle="Toggle Hydra">
        <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="size-6">
          <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" />
          <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
        </svg>
      </Button>
    </div>
  );
};
