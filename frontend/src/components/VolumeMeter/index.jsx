import React, {
  useState,
  useEffect,
  useRef,
  createRef,
} from 'react';

const VolumeMeter = ({ subscribeToSampleBuffer }) => {
  const meterRef = createRef();

  const audioContextRef = useRef();
  const analyserRef = useRef();

  useEffect(() => {
    const audioContext = new AudioContext();
    audioContextRef.current = audioContext;
    analyserRef.current = audioContext.createAnalyser();

    const dataArray = new Uint8Array(analyserRef.current.frequencyBinCount);

    let stop = false;

    let lastVolume = 0;
    const calculateVolume = () => {
      if (stop) {
        return;
      }
      analyserRef.current.getByteFrequencyData(dataArray);

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
      // free resources
      audioContextRef.current.close();
    };
  }, [])

  useEffect(() => {
    const updateStream = (samples) => {
      const samplesChannel0 = Float32Array.from(samples[0]);
      const samplesChannel1 = Float32Array.from(samples[1]);

      const buffer = audioContextRef.current.createBuffer(2, samples.length, 44100);
      buffer.copyToChannel(samplesChannel0, 0);
      buffer.copyToChannel(samplesChannel1, 1);

      const source = audioContextRef.current.createBufferSource();
      source.buffer = buffer;
      source.connect(analyserRef.current);
      source.start();
    };

    const unsubscribe = subscribeToSampleBuffer(updateStream);

    return () => unsubscribe();
  }, [subscribeToSampleBuffer]);

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
