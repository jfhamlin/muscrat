import {
  createRef,
  useEffect,
  useState,
} from 'react';


export default () => {
  const [startTime, setStartTime] = useState();

  const timeRef = createRef();

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

      const time = new Date(new Date() - new Date(startTime));
      const minutes = time.getMinutes();
      const seconds = time.getSeconds();
      timeRef.current.innerHTML = `${minutes}:${seconds < 10 ? `0${seconds}` : seconds}`;
    };

    update();
    const interval = setInterval(update, 1000);

    return () => clearInterval(interval);
  }, [startTime]);

  return (
    <div>
      <div ref={timeRef} />
      <button
        onClick={() => {
          setStartTime(new Date());
        }}>{startTime ? 'Reset' : 'Start'}</button>
      <button
        onClick={() => {
          setStartTime();
        }}>Stop</button>
    </div>
  );
};
