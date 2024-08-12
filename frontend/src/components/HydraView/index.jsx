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

const CerberusView = () => {
  const [cerb, setCerb] = useState(null);
  const [canvas, setCanvas] = useState(null);

  useEffect(() => {
    if (!canvas) {
      return;
    }

    const cerb = new Cerberus({
      canvas,
    });

    setCerb(cerb);

    testCerb(cerb);

    return () => cerb.dispose();

  }, [canvas]);

  return (
    <>
      <canvas className="w-full h-full bg-white"
              ref={setCanvas} />
    </>
  );
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
    testCerb(cerb);
    cerb.setResolution(window.innerWidth, window.innerHeight);
    h.cerberus = cerb;
  }, []);

  useEffect(() => {
    return window.addEventListener("resize", (e)=>{
      if (!hydra) {
        return;
      }
      // set canvas size to window size
      hydra.setResolution(window.innerWidth, window.innerHeight);
      hydra.cerberus.setResolution(window.innerWidth, window.innerHeight);
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
      console.log(hydra.cerberus.getCanvas());
      synth.s1.init({
        src: hydra.cerberus.getCanvas(),
      });

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
    {/* <div>
        <CerberusView />
        </div> */}
    <canvas className="w-full h-full bg-black"
            ref={setCanvas} />
  </div>;
};
