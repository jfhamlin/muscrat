import React, {
  useEffect,
  useRef,
} from 'react';

import Heading from '../Heading';
import { SpectrogramProps } from '../../types';

interface Color {
  r: number;
  g: number;
  b: number;
}

class Gradient {
  private colors: Color[];

  constructor(colors: Color[]) {
    this.colors = colors;
  }

  getColor(fraction: number): string {
    const index = Math.floor(fraction * (this.colors.length - 1));
    if (index === this.colors.length - 1) {
      return this.mixColors(this.colors[index], this.colors[index], 0);
    }
    const startColor = this.colors[index];
    const endColor = this.colors[index + 1];
    const fractionOfColor = (fraction - (index / (this.colors.length - 1))) * (this.colors.length - 1);
    return this.mixColors(startColor, endColor, fractionOfColor);
  }

  private mixColors(startColor: Color, endColor: Color, fraction: number): string {
    const r = startColor.r + (endColor.r - startColor.r) * fraction;
    const g = startColor.g + (endColor.g - startColor.g) * fraction;
    const b = startColor.b + (endColor.b - startColor.b) * fraction;
    return `rgb(${r}, ${g}, ${b})`;
  }
}

// Available gradients for potential future use:
// const GRADIENT_BW = new Gradient([{ r: 0, g: 0, b: 0 }, { r: 255, g: 255, b: 255 }]);
// const GRADIENT_COOL = new Gradient([{ r: 0, g: 0, b: 128 }, { r: 0, g: 255, b: 255 }, { r: 0, g: 255, b: 128 }, { r: 64, g: 224, b: 208 }]);

const GRADIENT_INFRARED = new Gradient([
  { r: 0, g: 0, b: 0 },
  { r: 255, g: 0, b: 0 },
  { r: 255, g: 255, b: 0 },
  { r: 255, g: 255, b: 255 },
]);

const gradient = GRADIENT_INFRARED;

const Spectrogram: React.FC<SpectrogramProps> = ({ analyser, sampleRate }) => {
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (!analyser) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // clear the canvas with gradient 0 value
    ctx.fillStyle = gradient.getColor(0);
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    const bufferLength = analyser.frequencyBinCount;
    const dataArray = new Uint8Array(bufferLength);

    const nyquist = sampleRate / 2;
    const frequencyResolution = nyquist / bufferLength;
    const binCenterFrequencies = Array.from({ length: bufferLength }, (_, i) => i * frequencyResolution + frequencyResolution / 2);

    // in log2 scale, the normalized [0, 1] x positions of the frequency bins,
    // with 20hz to 20khz mapped to [0, 1]
    const logNyquist = Math.log2(nyquist);
    // minLogFrequency is the log2 of the smallest bin > 20hz
    const minLogFrequency = Math.log2(binCenterFrequencies.filter(frequency => frequency > 20)[0]);
    const logBinPositions = binCenterFrequencies.map(frequency => (Math.log2(frequency) - minLogFrequency) / (logNyquist - minLogFrequency));

    let stop = false;

    const renderFrame = (): void => {
      if (stop) return;

      requestAnimationFrame(renderFrame);

      // set canvas width and height to be the same as its CSS size
      const rect = canvas.getBoundingClientRect();
      if (canvas.width !== rect.width || canvas.height !== rect.height) {
        canvas.width = rect.width;
        canvas.height = rect.height;
        // clear the canvas with gradient 0 value
        ctx.fillStyle = gradient.getColor(0);
        ctx.fillRect(0, 0, canvas.width, canvas.height);
      }
      const width = canvas.width;
      const height = canvas.height;

      // Get the frequency data
      analyser.getByteFrequencyData(dataArray);

      // Scroll the canvas
      const imageData = ctx.getImageData(0, 0, width, height);
      ctx.putImageData(imageData, 0, -1);

      if (false) {
        // Draw the new line at the bottom, in log2 scale
        dataArray.forEach((value, i) => {
          const percent = value / 255;
          // const y = height - 1 - Math.floor((height - 1) * percent); // Draw from bottom (unused)
          const x = logBinPositions[i] * width;
          let x2 = width;
          if (i < bufferLength - 1) {
            x2 = logBinPositions[i + 1] * width;
          }

          ctx!.fillStyle = '#000';
          ctx!.fillStyle = gradient.getColor(percent);

          ctx!.fillRect(x, height - 1, x2, 1); // Draw single pixel line
        });
      } else {
        const BLOCK_SIZE = 4;
        const BLOCK_FRAC = BLOCK_SIZE / width;
        // Draw four pixels wide at a time in log2 scale, averaging the values
        // from all bins covered
        let curBin = 0; // start bin search here
        for (let x = 0; x < width; x += BLOCK_SIZE) {
          const logX = x / width;
          while (curBin < bufferLength - 1 && logBinPositions[curBin + 1] < logX) {
            ++curBin;
          }
          let sum = 0;
          let count = 0;
          // A very simple average of the values in the bins covered, not weighted
          // by the fraction of the bin covered.
          for (let i = curBin; i < bufferLength && logBinPositions[i] < logX + BLOCK_FRAC; ++i) {
            sum += dataArray[i];
            ++count;
          }

          const avg = count > 0 ? sum / count / 255 : 0;
          ctx!.fillStyle = gradient.getColor(avg);
          ctx!.fillRect(x, height - 1, BLOCK_SIZE, 1); // Draw four pixels wide
        }
      }
    };

    requestAnimationFrame(renderFrame);

    return () => {
      stop = true;
    };
  }, [analyser, sampleRate]);

  // round the corners of the canvas

  return (
    <div className="w-full h-full relative">
      <Heading>
        <h1>Frequency Spectrum</h1>
      </Heading>
      <canvas className="w-full h-full rounded-lg"
              ref={canvasRef} />
    </div>
  );
};

export default Spectrogram;