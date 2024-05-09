import {
  useState,
  useEffect,
} from 'react';

import {
  GetKnobs,
} from "../../../wailsjs/go/main/App";

import {
  EventsOn,
  EventsEmit,
} from '../../../wailsjs/runtime';

const Knob = ({ knob }) => {
  const [value, setValue] = useState(knob.def);

  return (
    <div>
      <h2>{knob.name}</h2>
      {/* input, focus on click */}
      <input
        type="range"
        min={knob.min}
        max={knob.max}
        step={knob.step ?? 0.1}
        value={value}
        onChange={(e) => {
          EventsEmit('knob-value-change', knob.id, new Number(e.target.value));
          setValue(e.target.value);
        }}
        onClick={(e) => {
          e.target.focus();
        }}
      />
      {/* value */}
      <div>{value}</div>
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

  return (
    <div className="mx-2 my-2">
      <h1>Knobs</h1>
      {knobs.map((knob) => (
        <Knob key={knob.id} knob={knob} />
      ))}
    </div>
  )
}
