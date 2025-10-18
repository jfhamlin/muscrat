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

const Knob = ({ knob, color }) => {
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
              color={color}
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

const GROUP_COLORS = [
  '#ef4444',
  '#f97316',
  '#eab308',
  '#22c55e',
  '#06b6d4',
  '#3b82f6',
  '#a855f7',
  '#ec4899',
];

const hashString = (str) => {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash;
  }
  return Math.abs(hash);
};

const getGroupColor = (groupName) => {
  const hash = hashString(groupName);
  return GROUP_COLORS[hash % GROUP_COLORS.length];
};

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

const groupKnobs = (knobs) => {
  const groups = {};
  const ungrouped = [];

  knobs.forEach((knob) => {
    if (knob.group && knob.group !== '') {
      if (!groups[knob.group]) {
        groups[knob.group] = [];
      }
      groups[knob.group].push(knob);
    } else {
      ungrouped.push(knob);
    }
  });

  return { groups, ungrouped };
}

const KnobGroup = ({ groupName, knobs }) => {
  const color = getGroupColor(groupName);
  return (
    <div className="border rounded p-2 relative" style={{ borderColor: color }}>
      <div className="absolute top-1 left-2 text-xs font-medium" style={{ color }}>
        {groupName}
      </div>
      <div className="flex flex-wrap gap-2 mt-4">
        {knobs.map((knob) => (
          <Knob key={knob.id} knob={knob} color={color} />
        ))}
      </div>
    </div>
  );
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

  const { groups, ungrouped } = groupKnobs(knobs);
  const groupNames = Object.keys(groups).sort();

  return (
    <div className="mt-2 overflow-auto w-full h-full select-none">
      {knobs.length === 0 ? null :
       <div className="flex flex-wrap gap-2">
         {ungrouped.map((knob) => (
           <Knob key={knob.id} knob={knob} />
         ))}
         {groupNames.map((groupName) => (
           <KnobGroup key={groupName} groupName={groupName} knobs={groups[groupName]} />
         ))}
       </div>}
    </div>
  )
}
