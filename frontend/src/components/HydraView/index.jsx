import React, {
  useEffect,
  useState,
  useCallback,
} from 'react';

import Hydra from 'hydra-synth';

export default () => {
  const [hydra, setHydra] = useState(null);

  const setCanvas = useCallback((canvas) => {
    if (canvas) {
      setHydra(new Hydra({
        canvas,
        detectAudio: false,
        autoLoop: true,
      }));
    }
  }, []);

  useEffect(() => {
    if (!hydra) {
      return;
    }
    const synth = hydra.synth;
    synth.osc(10, 0.1, 0.35)
         .rotate(0.5*Math.PI)
         .out();
  }, [hydra]);

  return <>
    <canvas className="w-full h-full"
            ref={setCanvas} />
  </>;
};
