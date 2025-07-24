import {
  useState,
  useEffect,
} from 'react';


import {
  GetSampleRate,
} from "../../../bindings/github.com/jfhamlin/muscrat/muscratservice";
import { Events } from "@wailsio/runtime";

import Oscilloscope from "../Oscilloscope";
import VolumeMeter from "../VolumeMeter";
import Spectrogram from "../Spectrogram";
import Console from "../Console";
import Timer from "../Timer";
import Knobs from "../Knobs";
import TabBar from "../TabBar";
import ScopeView from "../ScopeView";

import logo from '../../assets/images/muscrat.svg';

const createAudioResources = () => {
  const audioContext = new AudioContext();
  const analyser = audioContext.createAnalyser();

  return {
    context: audioContext,
    analyser,
  };
};

const AudioVisualizers = () => {
  const [audioResources, setAudioResources] = useState(null);
  const [sampleRate, setSampleRate] = useState(44100);

  useEffect(() => {
    GetSampleRate().then((sampleRate) => {
      setSampleRate(sampleRate);
    });
  }, []);

  useEffect(() => {
    const audioResources = createAudioResources();
    setAudioResources(audioResources);

    let nextBufferTime = audioResources.context.currentTime;

    const unsubscribe = Events.On("samples", (evt) => {
      const samples = evt.data[0];
      const samplesChannel0 = Float32Array.from(samples[0]);
      const samplesChannel1 = Float32Array.from(samples[1]);

      const bufferLength = samplesChannel0.length;

      const context = audioResources.context;
      const analyser = audioResources.analyser;

      const buffer = context.createBuffer(2, bufferLength, sampleRate);
      buffer.copyToChannel(samplesChannel0, 0);
      buffer.copyToChannel(samplesChannel1, 1);

      const source = context.createBufferSource();
      source.buffer = buffer;
      source.connect(analyser);
      source.start(nextBufferTime);

      nextBufferTime += bufferLength / sampleRate;
    });

    return () => {
      unsubscribe();
      audioResources.context.close();
    };
  }, [sampleRate]);

  return (
    <div className="mb-1">
      <div className="h-40 mb-1">
        <Oscilloscope analyser={audioResources?.analyser} />
      </div>
      <div className="h-60">
        <Spectrogram analyser={audioResources?.analyser} sampleRate={sampleRate} />
      </div>
    </div>
  );
};

const old = () => {
  const [expanded, setExpanded] = useState(false);

  // For the timer
  const [startTime, setStartTime] = useState();

  const toggleExpanded = () => {
    setExpanded((expanded) => !expanded);
  };

  // header has the logo and an svg button to expand (or collapse) the body
  const header = (
    <div className="flex flex-row items-center justify-between w-full mt-4 select-none">
      <img src={logo} className="w-32 max-w-32 my-4" alt="logo" />
      <button onClick={toggleExpanded} className="p-2 focus:outline-none">
        {expanded ? (
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            className="w-6 h-6"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M6 18L18 6M6 6l12 12"
            />
          </svg>
        ) : (
          <svg
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            className="w-6 h-6"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth="2"
              d="M4 8h16M4 16h16"
            />
          </svg>
        )}
      </button>
    </div>
  );


  if (!expanded) {
    return (
      <div className="flex flex-col items-center h-full mx-1">
        {header}
        <div className="flex flex-col flex-grow overflow-hidden mb-1">
          <VolumeMeter />
        </div>
        <Timer startTime={startTime} setStartTime={setStartTime} />
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center h-full mx-1">
      {header}
      {/* <img src={logo} className="w-32 max-w-32 mt-8 my-4" alt="logo" /> */}
      <div className="flex flex-col w-96 flex-grow overflow-hidden mb-1">
        <AudioVisualizers />
        <div className="flex-grow overflow-auto">
          <Console />
        </div>
      </div>
    </div>
  );
};

import Knob from "../Knob";

export default () => {
  const [selectedTab, setSelectedTab] = useState("params");

  return (
    <div className="ml-2 mr-5 pt-[6px] h-full flex flex-col justify-between">
      <div className="flex-grow flex flex-col overflow-hidden">
        <TabBar options={["params", "graph"]}
                selected={selectedTab}
                onSelect={setSelectedTab} />
        <div className="flex-grow overflow-hidden relative">
          <div className={`absolute inset-0 ${selectedTab === "params" ? "block" : "hidden"}`}>
            <Knobs />
          </div>
          <div className={`absolute inset-0 ${selectedTab === "graph" ? "block" : "hidden"}`}>
            <ScopeView />
          </div>
        </div>
      </div>
      <div className="flex mt-2 mb-5 pt-2 border-t border-gray-300/25">
        <VolumeMeter />
      </div>
    </div>
  );
};
