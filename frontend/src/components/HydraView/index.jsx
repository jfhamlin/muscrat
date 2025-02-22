import React, {
  useEffect,
  useState,
  useCallback,
  useRef,
} from 'react';

import { Events } from "@wailsio/runtime";

import Hydra from 'hydra-synth';

import Cerberus from '../../lib/cerberus';

const testCerb = (cerb) => {
  cerb.scene().add(
    cerb.dodecahedron().translate(2, 0, 0).color(0.2, 0, 1),
    cerb.sphere()
        .scale(0.5),
    cerb.box()
        .scale()
        .translate(2)
        .rotate(0, 0, 0, 0, 0, 0.1)
        .color(1, 0.2, ({ time }) => 0.5 + 0.5*Math.sin(Math.PI * time)),
    cerb.box()
        .translate(-2, 0, 0)
        .color(0, 0, 1)
        .rotate(0,
                ({ time }) => 0.1 * Math.PI * time,
                0),
    cerb.torus(0.5, 0.1)
        .color(1, 0, 0.1)
        .rotate(0, 0, 0, 0.5),
    cerb.torusKnot(1, 0.1, 128, 16, 5, 7)
        .color(0.1, 1, 0.1),
    cerb.pointLight(0xffffff, 200, 100)
        .translate(5, 5, 5),
    cerb.ambientLight(0x404040, 2)
        .color(({ time }) => 2*Math.sin(Math.PI * time), 0, 0),
  ).render(cerb.camera()
               .translate(0, 0, 5)
               .rotate(0, 0, 0, 0.01, 0.1, 0.01)
               .lookAt(2, 0, 0));
};

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

    const cerb = new Cerberus();
    cerb.setResolution(window.innerWidth, window.innerHeight);
    h.synth.cerberus = () => cerb;
  }, []);

  useEffect(() => {
    return window.addEventListener("resize", (e)=>{
      if (!hydra) {
        return;
      }
      // set canvas size to window size
      hydra.setResolution(window.innerWidth, window.innerHeight);
      hydra.synth.cerberus().setResolution(window.innerWidth, window.innerHeight);
    });
  }, [hydra]);

  useEffect(() => {
    return Events.On("hydra.expr", (evt) => {
      const data = evt.data[0];
      setExpr(data.expr);
      setVars(new Set(data.vars));
    });
  }, []);
  useEffect(() => {
    return Events.On("hydra.mapping", (evt) => {
      const mapping = evt.data[0];
      mappings.current = mapping;
    });
  }, []);

  useEffect(() => {
    if (!hydra) {
      return;
    }

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
        case "do":
          // special form, execute a list of expressions
          for (let i = 1; i < call.length; i++) {
            const res = evalExpr(self, call[i]);
            if (i === call.length - 1) {
              // return the last result
              return res;
            }
          }
        default:
          const args = call.slice(1).map((arg) => evalExpr(synth, arg));
          console.log("calling", call[0], "with", args);
          return self[call[0]].apply(self, args);
      }
    };
    evalExpr = (self, expr) => {
      // if it's an array, it's a function
      if (Array.isArray(expr)) {
        return evalCall(self, expr);
      }
      // if it's a map, evaluate each value
      if (typeof expr === "object") {
        const evaluated = {};
        for (const [key, value] of Object.entries(expr)) {
          evaluated[key] = evalExpr(self, value);
        }
        return evaluated;
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
