import {
  useState,
  useEffect,
  useRef,
} from 'react';

import { EventsOn } from '../../../wailsjs/runtime';

const ErrorIcon = () => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
    </svg>
  );
};

const WarnIcon = () => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
    </svg>

  );
}

const COLORS = {
  debug: 'bg-gray-100',
  info: 'bg-blue-100',
  warn: 'bg-yellow-100',
  error: 'bg-red-100',
};

const Event = ({ event, count }) => {
  const level = event.level; // debug, info, warn, error
  const message = event.message;
  const data = event.data;

  const [dataVisible, setDataVisible] = useState(false);
  const toggleData = () => setDataVisible((prev) => !prev);
  const dataElement = data && (
    <div className="text-gray-500">
      <pre>{data}</pre>
    </div>
  );

  let dataRender = data;
  if (typeof data === 'object') {
    dataRender = JSON.stringify(data, null, 2);
  }

  let icon;
  switch (level) {
    case 'warn':
      icon = <WarnIcon />;
      break;
    case 'error':
      icon = <ErrorIcon />;
      break;
  }

  let bgColor = COLORS[level] || '';

  return (
    <div className={"text-xs p-2 flex items-center rounded " + bgColor}>
      {count > 1 && <div className="mr-2 text-gray-500">{count}x</div>}
      {icon && <div className="mr-2">{icon}</div>}
      <div className="flex-1">
        <div>{message}</div>
        {dataVisible && dataElement}
      </div>
    </div>
  );
};

export default () => {
  const [events, setEvents] = useState([]);

  const ref = useRef(null);

  useEffect(() => {
    return EventsOn('console.log', (data) => {
      setEvents((prev) => {
        const last = prev[prev.length - 1];
        if (last && JSON.stringify(last.event) === JSON.stringify(data)) {
          return [...prev.slice(0, -1), {event: data, count: last.count + 1}];
        }
        return [...prev, {event: data, count: 1}];
      });

      // scroll to bottom if already at bottom

      if (ref.current && ref.current.scrollHeight - ref.current.scrollTop === ref.current.clientHeight) {
        console.log('scrolling to bottom');
        requestAnimationFrame(() => {
          console.log('scrolling to bottom !!');
          ref.current.scrollTop = ref.current.scrollHeight;
        });
      }
    });
  }, []);

  return (
    <div ref={ref} className="bg-white rounded-lg flex-grow overflow-auto">
      {events.map(({ event, count }, i) => (
        <Event key={i} event={event} count={count} />
      ))}
    </div>
  )
}
