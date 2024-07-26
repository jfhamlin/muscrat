import React, {
  useEffect,
  useState,
  useCallback,
  useRef,
} from 'react';

import {
  EventsOn,
  EventsEmit,
} from '../../../wailsjs/runtime';

import Hydra from 'hydra-synth';

export default () => {
  const [hydra, setHydra] = useState(null);

  const [expr, setExpr] = useState({render: ["solid"]});
  const [vars, setVars] = useState(new Set())

  const [canvasSize, setCanvasSize] = useState();

  const mappings = useRef({});

  const setCanvas = useCallback((canvas) => {
    if (canvas) {
      setCanvasSize([canvas.width, canvas.height]);
      setHydra(new Hydra({
        canvas,
        detectAudio: false,
        autoLoop: true,
      }));
    }
  }, []);

  useEffect(() => {
    return EventsOn("hydra.expr", (expr) => {
      setExpr(expr.expr);
      setVars(new Set(expr.vars));
    });
  }, []);
  useEffect(() => {
    return EventsOn("hydra.mapping", (mapping) => {
      mappings.current = mapping;
    });
  }, []);

  useEffect(() => {
    if (!hydra || !canvasSize) {
      return;
    }

    console.log(hydra);
    hydra.synth.setResolution(4*canvasSize[0], 4*canvasSize[1]);
  }, [hydra, canvasSize]);

  useEffect(() => {
    if (!hydra) {
      return;
    }

    console.log("new hydra expr", expr);
    const synth = hydra.synth;

    let evalMapping, evalCall, evalExpr;

    evalMapping = (name) => {
      if (vars.has(name)) {
        return () => mappings.current[name] ?? 0;
      }
      return synth[name];
    };
    evalCall = (self, call) => {
      switch (call[0]) {
        case "__lookup":
          return evalMapping(call[1]);
        case "..":
          // special form, chaining method calls
          let chained = evalExpr(self, call[1]);
          for (let i = 2; i < call.length; i++) {
            chained = evalCall(chained, call[i]);
          }
          return chained;
        default:
          return self[call[0]].apply(self, call.slice(1).map((arg) => evalExpr(synth, arg)));
      }
    };
    evalExpr = (self, expr) => {
      // if it's an array, it's a function
      if (Array.isArray(expr)) {
        return evalCall(self, expr);
      }
      // otherwise, it's a constant
      return expr;
    };

    try {
      const sources = expr.sources ?? {};
      for (const [name, source] of Object.entries(sources)) {
        evalExpr(synth[name], source);
      }
      evalExpr(synth, expr.render);
    } catch (e) {
      EventsEmit("console.log", {
        level: "error",
        message: e.message,
        data: e.stack,
      });
      console.error("error setting expr", e);
    }
  }, [hydra, expr]);

  return <>
    <canvas className="w-full h-full bg-black"
            ref={setCanvas} />
  </>;
};
