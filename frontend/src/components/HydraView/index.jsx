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

  const [graph, setGraph] = useState(["solid"]);

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
      synth.stop();
      return;
    }

    let evalMapping, evalCall, evalExpr;

    evalMapping = (name) => {
      // if mappings has the key, use it, otherwise use the synth
      if (mappings.current[name]) {
        return () => mappings.current[name] ?? 0;
      }
      return synth[name];
    };
    evalCall = (self, call) => {
      switch (call[0]) {
        case "..":
          // special form, chaining method calls
          let chained = evalExpr(self, call[1]);
          for (let i = 2; i < call.length; i++) {
            chained = evalCall(chained, call[i]);
          }
          return chained;
        default:
          console.log("calling", call[0]);
          return self[call[0]].apply(self, call.slice(1).map((arg) => evalExpr(synth, arg)));
      }
    };
    evalExpr = (self, expr) => {
      // if it's an array, it's a function
      if (Array.isArray(expr)) {
        return evalCall(self, expr);
      }
      // if it's a string, it's a mapping
      if (typeof expr === "string") {
        return evalMapping(expr);
      }
      // otherwise, it's a constant
      return expr;
    };

    try {
      evalExpr(synth, graph);
    } catch (e) {
      console.error("error setting graph", e);
    }
  }, [hydra, graph]);

  return <>
    <canvas className="w-full h-full"
            ref={setCanvas} />
  </>;
};
