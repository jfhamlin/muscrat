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

const gradient = GRADIENT_COOL;

export default ({ analyser, sampleRate }) => {
  const canvasRef = createRef();

  useEffect(() => {
    if (!analyser) return;

    const canvas = canvasRef.current;
    const ctx = canvas.getContext('2d');
    const bufferLength = analyser.frequencyBinCount
    const dataArray = new Uint8Array(bufferLength);

    // Calculate frequency resolution
    const frequencyResolution = sampleRate / 2 / bufferLength;

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
      const height = canvas.height - 20;

      // Get the frequency data
      analyser.getByteFrequencyData(dataArray);

      // Scroll the canvas
      const imageData = ctx.getImageData(0, 0, width, height);
      ctx.putImageData(imageData, 0, -1);

      // Draw the new line at the bottom
      dataArray.forEach((value, i) => {
        const percent = value / 255;
        const y = height - 1 - Math.floor((height - 1) * percent); // Draw from bottom
        const x = i * (width / bufferLength);

        ctx.fillStyle = '#000';
        ctx.fillStyle = gradient.getColor(percent)
        ctx.fillRect(x, height - 1, width / bufferLength, 1); // Draw single pixel line
      });

      // Update scroll position and eventually reset
      let newScrollY = scrollY + 1;
      if (newScrollY >= height) {
        newScrollY = 0;
      }

      // Draw frequency labels at the bottom
      if (scrollY % 20 === 0) { // Update frequency labels less frequently to make them readable
        ctx.fillStyle = '#000';
        ctx.fillRect(0, height, width, 20); // Cover old labels
        ctx.fillStyle = '#fff';
        ctx.font = '10px Arial';
        for (let i = 0; i < bufferLength; i += Math.round(bufferLength / 5)) {
          const frequency = (i * frequencyResolution).toFixed(0);
          ctx.fillText(`${frequency}Hz`, i * (width / bufferLength), height + 15);
        }
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
