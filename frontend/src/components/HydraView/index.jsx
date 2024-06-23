import React, {
  useEffect,
  useState,
  useCallback,
  useRef,
} from 'react';

import { EventsOn } from '../../../wailsjs/runtime';

import Hydra from 'hydra-synth';

export default () => {
  const [hydra, setHydra] = useState(null);

  const [graph, setGraph] = useState([]);

  const mappings = useRef({});

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
    return EventsOn("hydra.graph", (graph) => {
      setGraph(graph);
    });
  }, []);
  useEffect(() => {
    return EventsOn("hydra.mapping", (mapping) => {
      mappings.current = mapping;
    });
  }, []);

  useEffect(() => {
    if (!hydra) {
      return;
    }

    console.log("new hydra graph", graph);
    const synth = hydra.synth;
    if (!graph) {
      synth.out();
      return;
    }

    const processArg = (arg) => {
      // if a string, it refers to a mapping
      if (typeof arg === "string") {
        return () => mappings.current[arg] ?? 0;
      }
      return arg;
    };

    // graph will look something like
    // [["osc", 200, 0.5, 0], ["scrollX", 0.5, 1], ["add", "o0"], ["contrast", 10], ["color", (s) => 0.5*Math.sin(2*Math.PI*s.time)+1, 0.1, 1], ["out"]
    let node = synth;
    graph.reduce((prev, curr) => {
      node = node[curr[0]].apply(node, curr.slice(1).map(processArg));
    }, synth)
  }, [hydra, graph]);

  return <>
    <canvas className="w-full h-full"
            ref={setCanvas} />
  </>;
};
