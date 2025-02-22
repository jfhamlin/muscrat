import React, {
  useState,
  useEffect,
  createRef,
} from 'react';

import { Events } from "@wailsio/runtime";

const VolumeMeter = ({ }) => {
  const rmsMeterRefL = createRef();
  const rmsMeterRefR = createRef();
  const peakMeterRefL = createRef();
  const peakMeterRefR = createRef();

  useEffect(() => {
    const unsubscribe = Events.On('volume', (event) => {
      const data = event.data[0];
      const [rmsL, rmsR] = data.rms;
      const [peakL, peakR] = data.peak;

      if (!rmsMeterRefL.current) return;

      // set height
      rmsMeterRefL.current.style.height = `${Math.min(100, rmsL * 100)}%`;
      rmsMeterRefR.current.style.height = `${Math.min(100, rmsR * 100)}%`;

      peakMeterRefL.current.style.height = `${Math.min(100, peakL * 100)}%`;
      peakMeterRefR.current.style.height = `${Math.min(100, peakR * 100)}%`;

      // set color based on volume
      const getColor = (value) => {
        if (value < 0.6) {
          return 'green';
        } else if (value < 0.85) {
          return 'yellow';
        } else {
          return 'red';
        }
      }

      rmsMeterRefL.current.style.backgroundColor = getColor(rmsL);
      rmsMeterRefR.current.style.backgroundColor = getColor(rmsR);
      peakMeterRefL.current.style.backgroundColor = getColor(peakL);
      peakMeterRefR.current.style.backgroundColor = getColor(peakR);
    });
    return unsubscribe;
  }, [])

  const barColor = 'green';

  return (
    <div className="flex flex-row h-full items-end bg-gray-200">
      <div className="flex h-full items-end border border-black w-7">
        <div ref={rmsMeterRefL}
             className="border-t border-black w-full"
             style={{
               backgroundColor: barColor,
               transition: 'height 0.2s ease-out, background-color 0.1s ease-out' // Smooth transition for changes
             }}></div>
      </div>
      <div className="flex h-full items-end border border-black border-l-0 w-1">
        <div ref={peakMeterRefL}
             className="border-t border-black w-1"
             style={{
               backgroundColor: barColor,
               transition: 'height 0.2s ease-out, background-color 0.2s ease-out' // Smooth transition for changes
             }} />
      </div>
      <div className="flex h-full items-end border border-black border-l-0 w-7">
        <div ref={rmsMeterRefR}
             className="border-t border-black w-full"
             style={{
               backgroundColor: barColor,
               transition: 'height 0.2s ease-out, background-color 0.2s ease-out' // Smooth transition for changes
             }}></div>
      </div>
      <div className="flex h-full items-end border border-black border-l-0 w-1">
        <div ref={peakMeterRefR}
             className="border-t border-black w-1"
             style={{
               backgroundColor: barColor,
               transition: 'height 0.2s ease-out, background-color 0.2s ease-out' // Smooth transition for changes
             }} />
      </div>
    </div>
  );
};

export default VolumeMeter;
