import React, { useRef, useEffect, useState } from 'react';
import styles from './ScopeDisplay.module.css';

const ScopeDisplay = ({ samples, sampleRate, name, width = 400, height = 200 }) => {
  const canvasRef = useRef(null);
  const animationFrameRef = useRef(null);
  const [yScale, setYScale] = useState(1);
  const [timeScale, setTimeScale] = useState(1);
  const [triggerLevel, setTriggerLevel] = useState(0);
  const [frozen, setFrozen] = useState(false);
  const [lastSamples, setLastSamples] = useState(null);

  useEffect(() => {
    if (!frozen && samples && samples.length > 0) {
      setLastSamples(samples);
    }
  }, [samples, frozen]);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || !lastSamples || lastSamples.length === 0) return;

    const ctx = canvas.getContext('2d');
    const displaySamples = lastSamples;
    
    // Clear canvas
    ctx.fillStyle = '#1a1a1a';
    ctx.fillRect(0, 0, width, height);
    
    // Draw grid
    ctx.strokeStyle = '#333';
    ctx.lineWidth = 0.5;
    
    // Vertical grid lines (time divisions)
    const numTimeDivisions = 10;
    for (let i = 0; i <= numTimeDivisions; i++) {
      const x = (i / numTimeDivisions) * width;
      ctx.beginPath();
      ctx.moveTo(x, 0);
      ctx.lineTo(x, height);
      ctx.stroke();
    }
    
    // Horizontal grid lines (amplitude divisions)
    const numAmpDivisions = 8;
    for (let i = 0; i <= numAmpDivisions; i++) {
      const y = (i / numAmpDivisions) * height;
      ctx.beginPath();
      ctx.moveTo(0, y);
      ctx.lineTo(width, y);
      ctx.stroke();
    }
    
    // Draw center line
    ctx.strokeStyle = '#555';
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(0, height / 2);
    ctx.lineTo(width, height / 2);
    ctx.stroke();
    
    // Draw trigger level
    if (triggerLevel !== 0) {
      ctx.strokeStyle = '#ff0';
      ctx.lineWidth = 1;
      ctx.setLineDash([5, 5]);
      const triggerY = height / 2 - (triggerLevel * yScale * height / 2);
      ctx.beginPath();
      ctx.moveTo(0, triggerY);
      ctx.lineTo(width, triggerY);
      ctx.stroke();
      ctx.setLineDash([]);
    }
    
    // Draw waveform
    ctx.strokeStyle = '#0f0';
    ctx.lineWidth = 2;
    ctx.beginPath();
    
    const samplesToDisplay = Math.min(displaySamples.length, Math.floor(displaySamples.length * timeScale));
    const sampleStep = samplesToDisplay / width;
    
    for (let x = 0; x < width; x++) {
      const sampleIndex = Math.floor(x * sampleStep);
      if (sampleIndex < displaySamples.length) {
        const sample = displaySamples[sampleIndex];
        const y = height / 2 - (sample * yScale * height / 2);
        
        if (x === 0) {
          ctx.moveTo(x, y);
        } else {
          ctx.lineTo(x, y);
        }
      }
    }
    
    ctx.stroke();
    
    // Draw labels
    ctx.fillStyle = '#fff';
    ctx.font = '12px monospace';
    ctx.textAlign = 'left';
    ctx.fillText(name, 5, 15);
    
    // Draw scale info
    ctx.textAlign = 'right';
    const timePerDiv = (samplesToDisplay / sampleRate / numTimeDivisions * 1000).toFixed(2);
    ctx.fillText(`${timePerDiv}ms/div`, width - 5, height - 5);
    
    // Draw amplitude scale
    const ampPerDiv = (2 / numAmpDivisions / yScale).toFixed(2);
    ctx.fillText(`±${ampPerDiv}/div`, width - 5, 15);
    
  }, [lastSamples, width, height, yScale, timeScale, triggerLevel, name, sampleRate]);

  return (
    <div className={styles.scopeDisplay}>
      <canvas 
        ref={canvasRef} 
        width={width} 
        height={height}
        className={styles.canvas}
      />
      <div className={styles.controls}>
        <div className={styles.controlGroup}>
          <label>Y Scale</label>
          <input
            type="range"
            min="0.1"
            max="10"
            step="0.1"
            value={yScale}
            onChange={(e) => setYScale(parseFloat(e.target.value))}
          />
          <span>{yScale.toFixed(1)}x</span>
        </div>
        <div className={styles.controlGroup}>
          <label>Time Scale</label>
          <input
            type="range"
            min="0.1"
            max="1"
            step="0.01"
            value={timeScale}
            onChange={(e) => setTimeScale(parseFloat(e.target.value))}
          />
          <span>{(timeScale * 100).toFixed(0)}%</span>
        </div>
        <div className={styles.controlGroup}>
          <label>Trigger</label>
          <input
            type="range"
            min="-1"
            max="1"
            step="0.01"
            value={triggerLevel}
            onChange={(e) => setTriggerLevel(parseFloat(e.target.value))}
          />
          <span>{triggerLevel.toFixed(2)}</span>
        </div>
        <button 
          className={styles.freezeButton}
          onClick={() => setFrozen(!frozen)}
        >
          {frozen ? 'Run' : 'Freeze'}
        </button>
      </div>
    </div>
  );
};

export default ScopeDisplay;