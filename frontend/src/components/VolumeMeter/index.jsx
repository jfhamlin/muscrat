import React, {
  useState,
  useEffect,
  createRef,
} from 'react';

const VolumeMeter = ({ analyser }) => {
  const meterRef = createRef();

  useEffect(() => {
    if (!analyser) {
      return;
    }

    const dataArray = new Uint8Array(analyser.frequencyBinCount);

    let stop = false;

    let lastVolume = 0;
    const calculateVolume = () => {
      if (stop) {
        return;
      }
      analyser.getByteFrequencyData(dataArray);

      const sum = dataArray.reduce((acc, value) => acc + value * value, 0);
      const average = Math.sqrt(sum / dataArray.length);

      const updateCoeff = 0.2;
      const volume = (1 - updateCoeff) * lastVolume + updateCoeff * average;
      lastVolume = volume;

      // set height
      meterRef.current.style.height = `${Math.min(100, Math.max(0, 100 + volume - 100))}%`;

      requestAnimationFrame(calculateVolume);
    };

    requestAnimationFrame(calculateVolume);

    return () => {
      stop = true;
    };
  }, [analyser])

  const barColor = 'green';

  return (
    <div style={{ width: '20px', height: '200px', border: '1px solid #000', position: 'relative' }}>
      <div ref={meterRef}
        style={{
        position: 'absolute',
        bottom: 0,
        width: '100%',
        backgroundColor: barColor,
        transition: 'height 0.2s ease-out, background-color 0.2s ease-out' // Smooth transition for changes
      }}></div>
    </div>
  );
};

export default VolumeMeter;
