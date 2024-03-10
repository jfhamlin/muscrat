import {
  OpenFileDialog,
} from "../../../wailsjs/go/main/App";

import {
  useBuffersStore,
} from "../../contexts/buffers";

const Button = (props) => {
  const title = props.title;
  const onClick = props.onClick;
  return (
    <button onClick={onClick}>{title}</button>
  );
};

export default (props) => {
  const buffersStore = useBuffersStore();

  const handleLoadClick = () => {
    OpenFileDialog().then((response) => {
      console.log(response);
      const buffer = {
        fileName: response.FileName,
        content: response.Content,
      };
      buffersStore.addBuffer(buffer);
    }).catch((err) => {
      console.log(err);
    });
  };

  return (
    <div>
      <Button title="Load" onClick={handleLoadClick} />
    </div>
  );
};
