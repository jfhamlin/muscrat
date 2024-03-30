import {
  useState,
  useEffect,
  useRef,
} from 'react';

import { EventsOn } from '../../../wailsjs/runtime';

const Event = ({ event }) => {
  return (
    <div className="text-xs p-2">
      {JSON.stringify(event)}
    </div>
  );
};

export default () => {
  const [events, setEvents] = useState([]);

  const ref = useRef(null);

  useEffect(() => {
    return EventsOn('console.debug', (data) => {
      setEvents((prev) => [...prev, data]);

      // scroll to bottom if already at bottom
      if (ref.current && ref.current.scrollTop == ref.current.scrollHeight) {
        requestAnimationFrame(() => {
          ref.current.scrollTop = ref.current.scrollHeight;
        });
      }
    });
  }, []);

  return (
    <div className="bg-white rounded-lg p-2 h-full">
      <div ref={ref} className="overflow-y-auto overflow-x-hidden h-full">
        {events.map((event, i) => (
          <Event key={i} event={event} />
        ))}
      </div>
    </div>
  )
}
