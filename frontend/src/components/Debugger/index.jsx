import { useState } from 'react';

import Console from "../Console";
import Docs from "../Docs";
import TabBar from "../TabBar";

const Debugger = () => {
  const [selected, setSelected] = useState('console');

  const contentStyle = (contentName) => ({
    display: selected !== contentName ? "none" : undefined,
  });

  return (
    <div className="mx-5 flex-col h-full">
      <TabBar options={['console', 'docs']}
              selected={selected}
              onSelect={setSelected} />
      <div className="h-full overflow-hidden pb-5">
        <div className="h-full overflow-hidden" style={contentStyle('console')}>
          <Console />
        </div>
        <div className="h-full overflow-hidden" style={contentStyle('docs')}>
          <Docs />
        </div>
      </div>
    </div>
  );
};

export default Debugger;
