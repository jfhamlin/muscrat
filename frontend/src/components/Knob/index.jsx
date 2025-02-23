import {
  useState,
} from 'react';

const Knob = ({
  label,
  value,
  min,
  max,
  step,
  size,
  onChange,
}) => {
  return (
    <div style={{width: size}}>
      <div style={{
        width: size,
        height: size,
      }}>
        <div className="flex rounded-full p-4 w-full h-full border border-accent-primary items-center justify-center">
          <div className="flex items-center justify-center rounded-full bg-accent-primary/10"
               style={{
                 width: 0.4*size,
                 height: 0.4*size,
               }}>
            <div className="overflow-hidden text-ellipsis block text-accent-primary"
            style={{
              fontSize: 'clamp(0.25rem, 4vw, 2vw)',
            }}>
              {value}
            </div>
          </div>
        </div>
      </div>
      <div className="mt-1 text-center text-accent-primary">
        {label}
      </div>
    </div>
  );
};

export default Knob;

function CircularComponent() {
  return (
    <div className="flex flex-col items-center justify-center bg-[#1F1B2E] min-h-screen">
      {/* Outer container for the circular “ring” */}
      <div className="relative w-32 h-32 flex items-center justify-center text-[#F9BF5C]">
        {/* Conic gradient ring */}
        <div
          className="absolute inset-0 rounded-full border border-[#F9BF5C]"
          style={{
            /* 
              - conic-gradient(...) creates the wedge/different colors
              - maskImage (or WebkitMaskImage) carves out the center to form a ring
            */
            background: "conic-gradient(#8b6b57 0deg 100deg, #4e3b2d 100deg 360deg)",
            maskImage: "radial-gradient(circle, transparent 70%, white 71%)",
            WebkitMaskImage: "radial-gradient(circle, transparent 70%, white 71%)",
          }}
        />

        {/* Inner circle with the numeric value */}
        <div className="z-10 flex items-center justify-center w-16 h-16 rounded-full bg-[#1F1B2E] text-xl">
          33
        </div>
      </div>

      {/* Label below the circle */}
      <div className="mt-2 text-center text-[#F9BF5C]">
        low cutoff midi
      </div>
    </div>
  );
}
