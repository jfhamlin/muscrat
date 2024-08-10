import React, {
  useEffect,
  useState,
  useCallback,
  useRef,
} from 'react';

import { Events } from "@wailsio/runtime";

import Hydra from 'hydra-synth';

export default () => {
  const [hydra, setHydra] = useState(null);

  const [expr, setExpr] = useState({render: ["solid"]});
  const [vars, setVars] = useState(new Set())

  const mappings = useRef({});

  const setCanvas = useCallback((canvas) => {
    if (!canvas) {
      return;
    }
    const h = new Hydra({
      canvas,
      detectAudio: false,
      autoLoop: true,
    });
    h.setResolution(window.innerWidth, window.innerHeight);
    setHydra(h);
  }, []);

  useEffect(() => {
    return window.addEventListener("resize", (e)=>{
      if (!hydra) {
        return;
      }
      // set canvas size to window size
      hydra.setResolution(window.innerWidth, window.innerHeight);
    });
  }, [hydra]);

  useEffect(() => {
    return Events.On("hydra.expr", (evt) => {
      const data = evt.data;
      setExpr(data.expr);
      setVars(new Set(data.vars));
    });
  }, []);
  useEffect(() => {
    return Events.On("hydra.mapping", (evt) => {
      const mapping = evt.data;
      mappings.current = mapping;
    });
  }, []);

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
      Events.Emit("console.log", {
        level: "error",
        message: e.message,
        data: e.stack,
      });
      console.error("error setting expr", e);
    }
  }, [hydra, expr]);

  return <div className="w-full h-full">
    <canvas className="w-full h-full bg-black"
            ref={setCanvas} />
  </div>;
};
