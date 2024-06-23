import React, {
  useEffect,
  useState,
  useCallback,
  useRef,
} from 'react';

import { EventsOn } from '../../../wailsjs/runtime';

import Hydra from 'hydra-synth';

// updates resolution of canvas to match its size
const ResizableCanvas = ({ setCanvas, ...props }) => {
  const canvasRef = useRef(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) {
      return;
    }

    const resize = () => {
      const devicePixelRatio = window.devicePixelRatio || 1;
      canvas.width = devicePixelRatio * canvas.clientWidth;
      canvas.height = devicePixelRatio * canvas.clientHeight;
    };

    resize();
    window.addEventListener("resize", resize);
    return () => window.removeEventListener("resize", resize);
  }, []);

  useEffect(() => {
    setCanvas(canvasRef.current);
  }, [setCanvas]);

  return <canvas ref={canvasRef} {...props} />;
};

export default () => {
  const [hydra, setHydra] = useState(null);

  const [expr, setExpr] = useState(["solid"]);
  const [vars, setVars] = useState(new Set())

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
      // if it's a string, it's a mapping
      if (typeof expr === "string") {
        return evalMapping(expr);
      }
      // otherwise, it's a constant
      return expr;
    };

    try {
      evalExpr(synth, expr);
    } catch (e) {
      console.error("error setting expr", e);
    }
  }, [hydra, expr]);

  return <>
    <canvas className="w-full h-full bg-black"
            ref={setCanvas} />
  </>;
};
