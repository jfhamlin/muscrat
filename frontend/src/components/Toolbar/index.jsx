import {
  OpenFileDialog,
  PlayFile,
} from "../../../wailsjs/go/main/App";

import {
  useBuffersStore,
} from "../../contexts/buffers";

const Button = (props) => {
  const title = props.title;
  const onClick = props.onClick;
  let className = "bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded m-1"
  if (props.disabled) {
    className += " bg-gray-500 cursor-not-allowed";
  }
  return (
    <button className={className}
      {...props}>{title}</button>
  );
};

export default (props) => {
  const buffersStore = useBuffersStore();

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

  const handlePlayClick = () => {
    const selectedBufferName = buffersStore.selectedBufferName;
    const buffer = buffersStore.buffers[selectedBufferName];
    PlayFile(buffer.fileName);
  };

  return (
    <div>
      <Button title="Load" onClick={handleLoadClick} />
      <Button title="Play"
              disabled={!buffersStore.selectedBufferName}
              onClick={handlePlayClick} />
    </div>
  );
};
