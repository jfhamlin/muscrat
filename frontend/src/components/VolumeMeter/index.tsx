import React, {
  useEffect,
  useRef,
} from 'react';

import { Events } from "@wailsio/runtime";
import { VolumeMeterProps, VolumeEvent } from '../../types';

const VolumeMeter: React.FC<VolumeMeterProps> = ({}) => {
  const rmsMeterRefL = useRef<HTMLDivElement>(null);
  const rmsMeterRefR = useRef<HTMLDivElement>(null);
  const peakMeterRefL = useRef<HTMLDivElement>(null);
  const peakMeterRefR = useRef<HTMLDivElement>(null);

  // Peak hold state
  const peakHoldL = useRef<number>(0);
  const peakHoldR = useRef<number>(0);
  const peakHoldTimerL = useRef<NodeJS.Timeout | null>(null);
  const peakHoldTimerR = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    const unsubscribe = Events.On('volume', (event: VolumeEvent) => {
      const data = event.data[0];
      const [, ] = data.rms || [0, 0]; // rmsL, rmsR not used in current implementation
      const [peakL, peakR] = data.peak || [0, 0];
      const [rmsDBL, rmsDBR] = data.rmsDB || [-60, -60];
      const [, ] = data.peakDB || [-60, -60]; // peakDBL, peakDBR not used in current implementation

      if (!rmsMeterRefL.current || !rmsMeterRefR.current) return;

      // Convert dB values to normalized 0-1 range for display
      // -60dB to 0dB mapped to 0-1
      const dbToNormalized = (db: number): number => {
        return Math.max(0, Math.min(1, (db + 60) / 60));
      };

      // Apply logarithmic curve for better visual response
      const applyDisplayCurve = (normalized: number): number => {
        // Square root expands lower values for better visibility
        return Math.pow(normalized, 0.5);
      };

      const rmsDisplayL = applyDisplayCurve(dbToNormalized(rmsDBL));
      const rmsDisplayR = applyDisplayCurve(dbToNormalized(rmsDBR));

      // update the colors of children based on the volume
      // when on, upper children are red, lower are green
      // when off, they're black
      // a child is on if its normalized [0, 1] index is less than the volume

      const updateColor = (volume: number, children: HTMLCollection): void => {
        // children is an HTMLCollection, not an array

        for (let i = 0; i < children.length; i++) {
          const child = children[i] as HTMLElement;
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

      updateColor(rmsDisplayL, rmsMeterRefL.current.children);
      updateColor(rmsDisplayR, rmsMeterRefR.current.children);

      // Peak hold logic
      const updatePeakHold = (
        currentPeak: number,
        peakHoldRef: React.MutableRefObject<number>,
        peakHoldTimerRef: React.MutableRefObject<NodeJS.Timeout | null>,
        peakMeterRef: React.RefObject<HTMLDivElement>
      ): void => {
        if (currentPeak >= 1) {
          // Set peak hold
          peakHoldRef.current = 1;
          if (peakMeterRef.current) {
            peakMeterRef.current.style.backgroundColor = 'red';
          }

          // Clear existing timer
          if (peakHoldTimerRef.current) {
            clearTimeout(peakHoldTimerRef.current);
          }

          // Set new timer to clear peak after 2 seconds
          peakHoldTimerRef.current = setTimeout(() => {
            peakHoldRef.current = 0;
            if (peakMeterRef.current) {
              peakMeterRef.current.style.backgroundColor = 'black';
            }
          }, 2000);
        } else if (peakHoldRef.current === 0) {
          // Only update to black if not holding
          if (peakMeterRef.current) {
            peakMeterRef.current.style.backgroundColor = 'black';
          }
        }
      };

      updatePeakHold(peakL, peakHoldL, peakHoldTimerL, peakMeterRefL);
      updatePeakHold(peakR, peakHoldR, peakHoldTimerR, peakMeterRefR);
    });

    // Cleanup timers on unmount
    return () => {
      unsubscribe();
      if (peakHoldTimerL.current) clearTimeout(peakHoldTimerL.current);
      if (peakHoldTimerR.current) clearTimeout(peakHoldTimerR.current);
    };
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