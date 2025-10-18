import React, { useCallback } from 'react';
import { KnobProps } from '../../types';

const Knob: React.FC<KnobProps> = ({
  label,
  value,
  min,
  max,
  step: _step, // Renamed to indicate it's intentionally unused
  size,
  color,
  onChange,
}) => {
  size = size || 80;
  const knobColor = color || '#FFC36A';

  const gapAngle = 30;
  const fullAngle = 360 - gapAngle;
  const fullPercent = 100 * fullAngle / 360;

  // percent of the circle that the value represents
  const valuePercent = fullPercent * (value - min) / (max - min);

  // 0/360 is the top, 50% of value
  // 180 + gapAngle/2 is 0% of value
  // 180 - gapAngle/2 is 100% of value
  const angleToValue = (angle: number): number => {
    const minAngle = 180 + gapAngle / 2;
    if (angle < minAngle) {
      if (angle > 180) {
        angle = minAngle;
      } else {
        angle += 360;
      }
    }

    const val = min + (angle - minAngle) * (max - min) / fullAngle;
    // ignore step
    return Math.min(max, Math.max(min, val));
  };

  const calculateCoordinates = useCallback((clientX: number, clientY: number, container: HTMLElement) => {
    const rect = container.getBoundingClientRect();
    const centerX = rect.width / 2;
    const centerY = rect.height / 2;
    const clickX = clientX - rect.left;
    const clickY = clientY - rect.top;

    const offsetX = clickX - centerX;
    const offsetY = clickY - centerY;

    let angle = Math.atan2(offsetY, offsetX);
    angle = (angle * 180) / Math.PI + 90;
    angle = (angle + 360) % 360;

    const radius = Math.hypot(offsetX, offsetY);
    const normalizedRadius = radius / (rect.width / 2);

    return { angle, normalizedRadius, radius };
  }, []);


  const handleDragStart = useCallback((e: React.MouseEvent | React.TouchEvent) => {
    const container = e.currentTarget as HTMLElement;
    const isTouch = e.type === 'touchstart';
    const clientX = isTouch ? (e as React.TouchEvent).touches[0].clientX : (e as React.MouseEvent).clientX;
    const clientY = isTouch ? (e as React.TouchEvent).touches[0].clientY : (e as React.MouseEvent).clientY;

    const { angle } = calculateCoordinates(
      clientX,
      clientY,
      container
    );

    let trackedAngle = angle;

    onChange?.(angleToValue(trackedAngle));

    const handleDrag = (moveEvent: MouseEvent | TouchEvent) => {
      const moveClientX = isTouch ? (moveEvent as TouchEvent).touches[0].clientX : (moveEvent as MouseEvent).clientX;
      const moveClientY = isTouch ? (moveEvent as TouchEvent).touches[0].clientY : (moveEvent as MouseEvent).clientY;

      const result = calculateCoordinates(
        moveClientX,
        moveClientY,
        container
      );

      trackedAngle = result.angle;

      onChange?.(angleToValue(trackedAngle));
    };

    const handleDragEnd = () => {
      onChange?.(angleToValue(trackedAngle));

      document.removeEventListener('mousemove', handleDrag);
      document.removeEventListener('touchmove', handleDrag);
      document.removeEventListener('mouseup', handleDragEnd);
      document.removeEventListener('touchend', handleDragEnd);
    };

    document.addEventListener('mousemove', handleDrag);
    document.addEventListener('touchmove', handleDrag);
    document.addEventListener('mouseup', handleDragEnd);
    document.addEventListener('touchend', handleDragEnd);
  }, [calculateCoordinates]);

  return (
    <div style={{ width: size }}>
      <div style={{
        width: size,
        height: size,
      }}>
        <div className="flex rounded-full w-full h-full items-center justify-center relative"
             style={{ color: knobColor }}
             onMouseDown={handleDragStart}
        >
          <svg
            className="absolute inset-0 w-full h-full transform rotate-[105deg]"
            viewBox="0 0 100 100"
          >
            <SvgCircle
              radiusPercent={49}
              strokeOpacity="1"
              strokeWidth="1"
              strokePercent={fullPercent} />
            <SvgCircle
              radiusPercent={40}
              strokeOpacity="0.25"
              strokeWidth="19%"
              strokePercent={fullPercent} />
            <SvgCircle
              radiusPercent={40}
              strokeOpacity="0.75"
              strokeWidth="19%"
              strokePercent={valuePercent} />
          </svg>
          <div className="flex items-center justify-center rounded-full"
               style={{
                 backgroundColor: `${knobColor}20`,
               }}
               style={{
                 width: '45%',
                 height: '45%',
               }}>
            <div className="overflow-hidden text-ellipsis block"
                 style={{ color: knobColor }}
                 style={{
                   fontSize: 'clamp(0.25rem, 1.5vw, 1rem)',
                 }}>
              {
                // truncate to two decimal places
                // don't include trailing zeros
                value.toFixed(2).replace(/\.?0*$/, '')
              }
            </div>
          </div>
        </div>
      </div>
      <div className="text-center text-sm" style={{ color: knobColor }}>
        {label}
      </div>
    </div>
  );
};

export default Knob;

interface SvgCircleProps {
  radiusPercent: number;
  strokeOpacity: string;
  strokeWidth: string;
  strokePercent: number;
}

const SvgCircle: React.FC<SvgCircleProps> = ({
  radiusPercent,
  strokeOpacity,
  strokeWidth,
  strokePercent,
}) => {
  const fullCircle = 2 * Math.PI * radiusPercent;
  const dashPercent = (strokePercent / 100) * fullCircle;
  const gapPercent = fullCircle - dashPercent;
  return (
    <circle
      cx="50"
      cy="50"
      r={`${radiusPercent}%`}
      fill="none"
      stroke="currentColor"
      strokeOpacity={strokeOpacity}
      strokeWidth={strokeWidth}
      strokeDasharray={`${dashPercent}% ${gapPercent}%`}
    />
  );
};
