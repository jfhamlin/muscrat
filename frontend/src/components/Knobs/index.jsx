import {
  useState,
  useEffect,
} from 'react';

import { Knob as PRKnob } from 'primereact/knob';
import { InputNumber } from 'primereact/inputnumber';

import {
  GetKnobs,
} from "../../../bindings/github.com/jfhamlin/muscrat/muscratservice";

import { Events } from "@wailsio/runtime";

const Knob = ({ knob }) => {
  const [value, setValue] = useState(knob.def);
  const [midiSub, setMidiSub] = useState(null);

  const updateKnobValue = (value) => {
    Events.Emit({
      name: 'knob-value-change',
      data: [knob.id, new Number(value)],
    });
    setValue(value);
  }

  const knobValueChange = (value) => {
    // at most 4 decimal places
    updateKnobValue(parseFloat(value.toFixed(4)));
  }

  const subscribeMidi = () => {
    // wait for midi message
    setMidiSub({waiting: true});
  }

  useEffect(() => {
    if (!midiSub) {
      return;
    }

    if (midiSub.waiting) {
      const sub = Events.On('midi', (evt) => {
        const data = evt.data;
        const message = data.message;
        if (message.type !== 'controlChange') {
          return;
        }
        const deviceId = data.deviceId;
        const { channel, controller, value } = message;
        setMidiSub({
          waiting: false,
          deviceId: deviceId,
          channel: channel,
          controller: controller,
          initialValue: value,
        });
      });
      return sub;
    } else {
      const sub = Events.On('midi', (evt) => {
        const data = evt.data;
        const message = data.message;
        if (message.type !== 'controlChange') {
          return;
        }
        const { channel, controller, value } = message;
        if (midiSub.channel !== channel || midiSub.controller !== controller) {
          return;
        }
        const newValue = knob.min + (value / 127) * (knob.max - knob.min);
        updateKnobValue(newValue);

        /* const diff = value - midiSub.initialValue;
         * const newValue = value + diff;
         * knobValueChange(newValue); */
      });
      return sub;
    }
  }, [midiSub]);

  // component should flash if waiting for midi

  // label is centered
  return (
    <div className="border border-primary p-2 m-2 noscroll overflow-hidden">
      <h2 className="font-bold text-center text-black">
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
                     maxFractionDigits={4}
                     onValueChange={(e) => knobValueChange(e.value)} />
      </div>
      <button className={"bg-primary text-white p-1 m-1" + (midiSub?.waiting ? " animate-pulse" : "")}
              onClick={subscribeMidi}>
        MIDI
      </button>
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
    return Events.On('knobs-changed', (evt) => {
      const data = evt.data;
      sortKnobs(data);
      setKnobs(data);
    });
  }, []);

  return (
    <div className="mx-2 my-2 overflow-hidden w-full h-full">
      <h1 className="font-bold text-xl text-center fixed w-full left-0 top-0 my-1">
        Knobs
      </h1>
      <div className="mt-5 overflow-auto">
        {knobs.length === 0 ? null :
         <div className="flex flex-wrap justify-center">
           {knobs.map((knob) => (
             <Knob key={knob.id} knob={knob} />
           ))}
         </div>}
      </div>
    </div>
  )
}
