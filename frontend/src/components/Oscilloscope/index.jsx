import React, {
  useEffect,
  createRef,
} from 'react';

import Heading from '../Heading';

export default ({ analyser }) => {
  const canvasRef = createRef();

  useEffect(() => {
    if (!analyser) return;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext('2d');
    const bufferLength = analyser.fftSize;
    const dataArray = new Float32Array(bufferLength);

    const TARGET_FPS = 10;
    let lastTime = performance.now();

    let stop = false;
    const renderFrame = () => {
      if (stop) return;

      requestAnimationFrame(renderFrame);

      const currentTime = performance.now();
      if (currentTime - lastTime < 1000 / TARGET_FPS) {
        return;
      }
      lastTime = currentTime;

      // set canvas width and height to be the same as its CSS size
      const rect = canvas.getBoundingClientRect();
      if (canvas.width !== rect.width || canvas.height !== rect.height) {
        canvas.width = rect.width;
        canvas.height = rect.height;
      }

      const width = canvas.width;
      const height = canvas.height;

      // Clear the canvas
      ctx.fillStyle = 'rgb(0, 0, 0)';
      ctx.fillRect(0, 0, width, height);

      // draw horizontal line
      ctx.strokeStyle = 'rgb(128, 128, 128)';
      ctx.lineWidth = 0.5;
      ctx.beginPath();
      ctx.moveTo(0, height/2);
      ctx.lineTo(width, height/2);
      ctx.stroke();

      // Begin drawing the waveform
      ctx.lineWidth = 0.75;
      ctx.strokeStyle = 'rgb(255, 255, 255)';
      ctx.beginPath();

      const sliceWidth = width * 1.0 / (bufferLength - 1);
      let x = 0;

      // Fetch the time-domain data
      analyser.getFloatTimeDomainData(dataArray);

      for(let i = 0; i < bufferLength; i++) {
        const v = dataArray[i] * 0.5 + 0.5; // Normalize and center the waveform
        const y = v * height;

        if(i === 0) {
          ctx.moveTo(x, y);
        } else {
          ctx.lineTo(x, y);
        }

        x += sliceWidth;
      }

      ctx.stroke();
    };

    requestAnimationFrame(renderFrame);

    return () => {
      stop = true;
    };
  }, [analyser]);

  return (
    <div className="w-full h-full relative">
      <Heading>
        <h1>Oscilloscope</h1>
      </Heading>
      <canvas className="w-full h-full rounded-lg"
              ref={canvasRef} />
    </div>
  );
};
