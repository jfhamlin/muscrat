import React, {
  createRef,
  useEffect,
  RefObject,
} from 'react';
import { TimerProps } from '../../types';

const Timer: React.FC<TimerProps> = ({ startTime, setStartTime }) => {
  const timeRef: RefObject<HTMLDivElement> = createRef();

  // display the time since the start time
  // in format MM:SS
  useEffect(() => {
    const update = () => {
      if (!timeRef.current) {
        return;
      }

      if (!startTime) {
        timeRef.current.innerHTML = '-:-';
        return;
      }

      const time = new Date(new Date().getTime() - new Date(startTime).getTime());
      const minutes = time.getMinutes();
      const seconds = time.getSeconds();
      timeRef.current.innerHTML = `${minutes}:${seconds < 10 ? `0${seconds}` : seconds}`;
    };

    update();
    const interval = setInterval(update, 1000);

    return () => clearInterval(interval);
  }, [startTime, timeRef]);

  const handleStart = (): void => {
    setStartTime(new Date());
  };

  const handleStop = (): void => {
    setStartTime();
  };

  return (
    <div className="select-none">
      <div ref={timeRef} />
      <button onClick={handleStart}>
        {startTime ? 'Reset' : 'Start'}
      </button>
      <button onClick={handleStop}>
        Stop
      </button>
    </div>
  );
};

export default Timer;