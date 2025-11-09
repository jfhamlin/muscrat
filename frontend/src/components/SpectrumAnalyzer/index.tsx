import React, {
  useEffect,
  useRef,
} from 'react';

interface SpectrumAnalyzerProps {
  analyser: AnalyserNode | null;
  sampleRate: number;
}

const SpectrumAnalyzer: React.FC<SpectrumAnalyzerProps> = ({ analyser, sampleRate }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (!analyser) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const bufferLength = analyser.frequencyBinCount;
    const dataArray = new Uint8Array(bufferLength);

    const nyquist = sampleRate / 2;
    const frequencyResolution = nyquist / bufferLength;

    let stop = false;

    const renderFrame = (): void => {
      if (stop) return;

      requestAnimationFrame(renderFrame);

      const width = 400;
      const height = 150;

      // Get the frequency data
      analyser.getByteFrequencyData(dataArray);

      // Clear the canvas
      ctx.fillStyle = 'rgb(0, 0, 0)';
      ctx.fillRect(0, 0, width, height);

      // Draw the spectrum bars
      const BAR_WIDTH = 3;
      const BAR_GAP = 1;
      const TOTAL_BAR_WIDTH = BAR_WIDTH + BAR_GAP;
      const numBars = Math.floor(width / TOTAL_BAR_WIDTH);

      // Minimum frequency to display (20 Hz)
      const minFreq = 20;
      const maxFreq = Math.min(20000, nyquist);

      for (let i = 0; i < numBars; i++) {
        const x = i * TOTAL_BAR_WIDTH;
        const x2 = x + BAR_WIDTH;

        // Map x position to frequency (logarithmic)
        const xNorm = x / width;
        const x2Norm = x2 / width;
        const freq1 = minFreq * Math.pow(maxFreq / minFreq, xNorm);
        const freq2 = minFreq * Math.pow(maxFreq / minFreq, x2Norm);

        // Convert frequencies to bin indices
        const bin1 = Math.floor(freq1 / frequencyResolution);
        const bin2 = Math.ceil(freq2 / frequencyResolution);

        // Average the frequency data for this range
        let sum = 0;
        let count = 0;
        for (let j = Math.max(0, bin1); j < Math.min(bufferLength, bin2); j++) {
          sum += dataArray[j];
          count++;
        }

        const avg = count > 0 ? sum / count : 0;
        const normalizedValue = avg / 255;
        const barHeight = normalizedValue * height;

        // Color gradient: green -> yellow -> red based on level
        let color;
        if (normalizedValue < 0.6) {
          // Green to yellow
          const t = normalizedValue / 0.6;
          const r = Math.floor(t * 255);
          const g = 255;
          color = `rgb(${r}, ${g}, 0)`;
        } else if (normalizedValue < 0.8) {
          // Yellow to orange
          const t = (normalizedValue - 0.6) / 0.2;
          const r = 255;
          const g = Math.floor(255 - t * 100);
          color = `rgb(${r}, ${g}, 0)`;
        } else {
          // Orange to red
          const t = (normalizedValue - 0.8) / 0.2;
          const r = 255;
          const g = Math.floor(155 * (1 - t));
          color = `rgb(${r}, ${g}, 0)`;
        }

        ctx.fillStyle = color;
        ctx.fillRect(x, height - barHeight, BAR_WIDTH, barHeight);
      }

      // Draw frequency grid lines and labels
      ctx.strokeStyle = 'rgba(128, 128, 128, 0.3)';
      ctx.fillStyle = 'rgba(200, 200, 200, 0.8)';
      ctx.font = '9px monospace';
      ctx.lineWidth = 1;

      const frequencies = [100, 1000, 10000];
      frequencies.forEach(freq => {
        if (freq >= minFreq && freq <= maxFreq) {
          // Map frequency to x position using same logarithmic scale
          const xNorm = Math.log(freq / minFreq) / Math.log(maxFreq / minFreq);
          const x = xNorm * width;

          ctx.beginPath();
          ctx.moveTo(x, 0);
          ctx.lineTo(x, height);
          ctx.stroke();

          const label = freq >= 1000 ? `${freq / 1000}k` : `${freq}`;
          ctx.fillText(label, x + 2, 10);
        }
      });
    };

    requestAnimationFrame(renderFrame);

    return () => {
      stop = true;
    };
  }, [analyser, sampleRate]);

  return (
    <canvas
      className="rounded-sm"
      width={400}
      height={150}
      ref={canvasRef}
    />
  );
};

export default SpectrumAnalyzer;
