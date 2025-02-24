import React, {
  useState,
  useEffect,
  useRef,
} from 'react';

import { Events } from "@wailsio/runtime";

const VolumeMeter = ({ }) => {
  const rmsMeterRefL = useRef();
  const rmsMeterRefR = useRef();
  const peakMeterRefL = useRef();
  const peakMeterRefR = useRef();

  useEffect(() => {
    const unsubscribe = Events.On('volume', (event) => {
      const data = event.data[0];
      const [rmsL, rmsR] = data.rms;
      const [peakL, peakR] = data.peak;

      if (!rmsMeterRefL.current) return;

      // update the colors of children based on the volume
      // when on, upper children are red, lower are green
      // when off, they're black
      // a child is on if its normalized [0, 1] index is less than the volume

      const updateColor = (volume, children) => {
        // children is an HTMLCollection, not an array

        for (let i = 0; i < children.length; i++) {
          const child = children[i];
          const childValue = (children.length - i) / children.length;
          if (childValue < volume) {
            let color = 'green';
            if (childValue > 0.6) {
              color = 'yellow';
            }
            if (childValue > 0.8) {
              color = 'red';
            }
            child.style.backgroundColor = color;
          } else {
            child.style.backgroundColor = 'black';
          }
        }
      }

      updateColor(rmsL, rmsMeterRefL.current.children);
      updateColor(rmsR, rmsMeterRefR.current.children);

      if (peakL >= 1) {
        peakMeterRefL.current.style.backgroundColor = 'red';
      } else {
        peakMeterRefL.current.style.backgroundColor = 'black';
      }

      if (peakR >= 1) {
        peakMeterRefR.current.style.backgroundColor = 'red';
      } else {
        peakMeterRefR.current.style.backgroundColor = 'black';
      }
    });
    return unsubscribe;
  }, [])


  const barColor = 'green';

  // each channel is a stack of numCells cells
  const numCells = 10;

  return (
    <div className="flex flex-row h-30 gap-2">
      <div>
        <div ref={peakMeterRefL}
             className="w-5 h-2 bg-gray-950 mb-[4px] rounded-sm" />
        <div className="w-5 flex flex-col gap-[2px]" ref={rmsMeterRefL}>
          {
            Array.from({ length: numCells }).map((_, i) => (
              <div key={i} className="h-3 w-full bg-gray-950 rounded-sm"></div>
            ))
          }
        </div>
      </div>
      <div>
        <div ref={peakMeterRefR}
             className="w-5 h-2 bg-gray-950 mb-[4px] rounded-sm" />
        <div className="w-5 flex flex-col gap-[2px]" ref={rmsMeterRefR}>
          {
            Array.from({ length: numCells }).map((_, i) => (
              <div key={i} className="h-3 w-full bg-gray-950 rounded-sm"></div>
            ))
          }
        </div>
      </div>
    </div>
  );

  // below is the old version of the component

  return (
    <div className="flex flex-row h-20 items-end">
      <div className="flex h-full rounded-t-xl items-end border border-black w-4 bg-white">
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
      <div className="flex h-full rounded-t-xl items-end border border-black border-l-0 w-4 bg-white">
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
