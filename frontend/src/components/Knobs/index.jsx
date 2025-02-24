import {
  useState,
  useEffect,
} from 'react';

import PRKnob from '../Knob';

import { Link } from 'lucide-react';

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
    if (midiSub) {
      setMidiSub(null);
      return;
    }
    // wait for midi message
    setMidiSub({waiting: true});
  }

  useEffect(() => {
    if (!midiSub) {
      return;
    }

    if (midiSub.waiting) {
      const sub = Events.On('midi', (evt) => {
        const data = evt.data[0];
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
        const data = evt.data[0];
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
    <div className="noscroll overflow-hidden select-none relative">
      <PRKnob label={knob.name}
              value={value}
              min={knob.min}
              max={knob.max}
              step={knob.step ?? 0.1}
              size={80}
              onChange={(val) => knobValueChange(val)} />
      <button className={"absolute -top-2 -left-2 bg-primary p-1 m-1" +
                          (midiSub?.waiting ? " animate-pulse" : "") +
                          (midiSub ? " text-accent-primary" : " text-accent-primary/25")}
              onClick={subscribeMidi}>
        <Link size={14} />
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
      const data = evt.data[0];
      sortKnobs(data);
      setKnobs(data);
    });
  }, []);

  return (
    <div className="mt-2 overflow-hidden w-full h-full select-none">
      <div className="overflow-auto">
        {knobs.length === 0 ? null :
         <div className="flex flex-wrap gap-2">
           {knobs.map((knob) => (
             <Knob key={knob.id} knob={knob} />
           ))}
         </div>}
      </div>
    </div>
  )
}
