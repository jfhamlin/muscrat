import React, {
  useEffect,
  createRef,
} from 'react';

class Gradient {
  constructor(colors) {
    this.colors = colors;
  }

  getColor(fraction) {
    const index = Math.floor(fraction * (this.colors.length - 1));
    if (index === this.colors.length - 1) {
      return this.colors[index];
    }
    const startColor = this.colors[index];
    const endColor = this.colors[index + 1];
    const fractionOfColor = (fraction - (index / (this.colors.length - 1))) * (this.colors.length - 1);
    return this.mixColors(startColor, endColor, fractionOfColor);
  }

  mixColors(startColor, endColor, fraction) {
    const r = startColor.r + (endColor.r - startColor.r) * fraction;
    const g = startColor.g + (endColor.g - startColor.g) * fraction;
    const b = startColor.b + (endColor.b - startColor.b) * fraction;
    return `rgb(${r}, ${g}, ${b})`;
  }
}

const GRADIENT_BW = new Gradient([
  { r: 0, g: 0, b: 0 },
  { r: 255, g: 255, b: 255 },
]);

const GRADIENT_INFRARED = new Gradient([
  { r: 0, g: 0, b: 0 },
  { r: 255, g: 0, b: 0 },
  { r: 255, g: 255, b: 0 },
  { r: 255, g: 255, b: 255 },
]);

const GRADIENT_COOL = new Gradient([
  { r: 0, g: 0, b: 128 },    // Darker Blue
  { r: 0, g: 255, b: 255 },  // Cyan
  { r: 0, g: 255, b: 128 },  // Spring Green
  { r: 64, g: 224, b: 208 }, // Lighter Turquoise
]);

const gradient = GRADIENT_INFRARED;

export default ({ analyser, sampleRate }) => {
  const canvasRef = createRef();

  useEffect(() => {
    if (!analyser) return;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext('2d');
    const bufferLength = analyser.frequencyBinCount
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

    let scrollY = 0;

    let stop = false;

    const renderFrame = () => {
      if (stop) return;

      requestAnimationFrame(renderFrame);

      // set canvas width and height to be the same as its CSS size
      const rect = canvas.getBoundingClientRect();
      if (canvas.width !== rect.width || canvas.height !== rect.height) {
        canvas.width = rect.width;
        canvas.height = rect.height;
      }
      const width = canvas.width;
      const height = canvas.height;

      // Get the frequency data
      analyser.getByteFrequencyData(dataArray);

      // Scroll the canvas
      const imageData = ctx.getImageData(0, 0, width, height);
      ctx.putImageData(imageData, 0, -1);

      // Draw the new line at the bottom, in log2 scale
      dataArray.forEach((value, i) => {
        const percent = value / 255;
        const y = height - 1 - Math.floor((height - 1) * percent); // Draw from bottom
        const x = logBinPositions[i] * width;
        let x2 = width;
        if (i < bufferLength - 1) {
          x2 = logBinPositions[i + 1] * width;
        }

        ctx.fillStyle = '#000';
        ctx.fillStyle = gradient.getColor(percent)
        ctx.fillRect(x, height - 1, x2, 1); // Draw single pixel line
      });

      // Update scroll position and eventually reset
      let newScrollY = scrollY + 1;
      if (newScrollY >= height) {
        newScrollY = 0;
      }

      // Update scroll position
      scrollY = newScrollY;
    };

    requestAnimationFrame(renderFrame);

    return () => {
      stop = true;
    };
  }, [analyser, sampleRate]);

  return (
    <div className="w-full h-full">
      <canvas className="w-full h-full"
              ref={canvasRef} />
    </div>
  );
};
