import {
  useState,
  useEffect,
} from 'react';

import { Knob as PRKnob } from 'primereact/knob';
import { InputNumber } from 'primereact/inputnumber';

import {
  GetKnobs,
} from "../../../wailsjs/go/main/App";

import {
  EventsOn,
  EventsEmit,
} from '../../../wailsjs/runtime';

const Knob = ({ knob }) => {
  const [value, setValue] = useState(knob.def);

  const knobValueChange = (value) => {
    // at most 4 decimal places
    value = parseFloat(value.toFixed(4));
    EventsEmit('knob-value-change', knob.id, new Number(value));
    setValue(value);
  }

  // label is centered
  return (
    <div className="border border-primary p-2 m-2 noscroll overflow-hidden">
      <h2 className="font-bold text-center">
        {knob.name}
      </h2>
      <PRKnob value={value}
              min={knob.min}
              max={knob.max}
              step={knob.step ?? 0.1}
              size={80}
              onChange={(e) => knobValueChange(e.value)} />
      <div className="w-20">
        <InputNumber value={value}
                     min={knob.min}
                     max={knob.max}
                     step={knob.step ?? 0.1}
                     onValueChange={(e) => knobValueChange(e.value)} />
      </div>
    </div>
  )
}

const sortKnobs = (knobs) => {
  return knobs.sort((a, b) => {
    if (a.name < b.name) {
      return -1;
    }
    if (a.name > b.name) {
      return 1;
    }
    return 0;
  });
}

export default () => {
  const [knobs, setKnobs] = useState([]);

  useEffect(() => {
    GetKnobs().then((data) => {
      sortKnobs(data);
      setKnobs(data);
    });
    EventsOn('knobs-changed', (data) => {
      sortKnobs(data);
      setKnobs(data);
    });
  }, []);

  const style = {};
  if (knobs.length === 0) {
    // upright text
    style.writingMode = 'vertical-rl';
    style.textOrientation = 'upright';
  }

  return (
    <div className="mx-2 my-2 overflow-auto" style={style}>
      <h1 className="font-bold text-xl text-center">
        Knobs
      </h1>
      {knobs.length === 0 ? null :
       <div className="flex flex-wrap justify-center w-60">
         {knobs.map((knob) => (
           <Knob key={knob.id} knob={knob} />
         ))}
       </div>}
    </div>
  )
}
