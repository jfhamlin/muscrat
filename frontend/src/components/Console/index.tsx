import React, {
  useState,
  useEffect,
  useRef,
} from 'react';

import { Events } from "@wailsio/runtime";

import {
  Trash2 as TrashIcon,
} from 'lucide-react';

import {
  ConsoleEvent,
  EventWithCount,
  EventProps,
  ClearButtonProps,
} from '../../types';

const ErrorIcon: React.FC = () => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
    </svg>
  );
};

const WarnIcon: React.FC = () => {
  return (
    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" className="w-6 h-6">
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
    </svg>

  );
}

const COLORS: Record<string, string> = {
  warn: 'bg-yellow-500',
  error: 'bg-red-500',
};

const Event: React.FC<EventProps> = ({ event, count }) => {
  const level = event.level; // debug, info, warn, error
  const message = event.message;
  const data = event.data;

  const [dataVisible, setDataVisible] = useState<boolean>(false);
  const toggleData = (): void => setDataVisible((prev) => !prev);
  const dataElement = data && (
    <div className="text-gray-100 text-xs p-1">
      <pre>{data}</pre>
    </div>
  );


  let icon: React.ReactNode;
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
    <>
      <div className={"text-xs text-gray-100 p-1 flex items-center " + bgColor}>
        {count > 1 && <div className="mr-2 text-gray-100">{count}x</div>}
        {icon && <div className="mr-2">{icon}</div>}
        <div className="flex-1 overflow-x-scroll" onClick={toggleData}>
          <div>{message}</div>
        </div>
        {/* horizontal line */}
      </div>
      {dataVisible && dataElement}
      <hr className="border-gray-400" />
    </>
  );
};

const ClearButton: React.FC<ClearButtonProps> = ({ onClick, size = 16 }) => {
  return (
    <button className='text-gray-200' onClick={onClick} title="Clear console">
      <TrashIcon size={size} />
    </button>
  );
};

interface WailsConsoleEvent {
  data: [ConsoleEvent];
}

const Console: React.FC = () => {
  const [events, setEvents] = useState<EventWithCount[]>([]);

  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const unsubscribe = Events.On('console.log', (evt: WailsConsoleEvent) => {
      const data = evt.data[0];

      setEvents((prev) => {
        const last = prev[prev.length - 1];
        if (last && JSON.stringify(last.event) === JSON.stringify(data)) {
          return [...prev.slice(0, -1), {event: data, count: last.count + 1}];
        }
        return [...prev, {event: data, count: 1}];
      });

      // scroll to bottom if already at bottom
      if (ref.current && ref.current.scrollHeight - ref.current.scrollTop === ref.current.clientHeight) {
        requestAnimationFrame(() => {
          if (ref.current) {
            ref.current.scrollTop = ref.current.scrollHeight;
          }
        });
      }
    });

    return () => unsubscribe();
  }, []);

  return (
    <div className="flex flex-col h-full flex-grow">
      <div className="px-1 pb-2 flex items-center justify-end border-b border-gray-600/25">
        <ClearButton onClick={() => setEvents([])} />
      </div>
      <div ref={ref} className="flex-grow overflow-auto">
        {events.map(({ event, count }, i) => (
          <Event key={i} event={event} count={count} />
        ))}
      </div>
    </div>
  )
}

export default Console;