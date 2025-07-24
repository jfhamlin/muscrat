import React, { useState, useEffect, useRef } from 'react';
import { Events } from "@wailsio/runtime";
import ScopeDisplay from './ScopeDisplay';
import styles from './index.module.css';

const ScopeView = () => {
  const [scopes, setScopes] = useState({});
  const [scopeList, setScopeList] = useState([]);
  const [containerWidth, setContainerWidth] = useState(800);
  const containerRef = useRef(null);

  // Handle container resize
  useEffect(() => {
    const updateWidth = () => {
      if (containerRef.current) {
        const width = containerRef.current.offsetWidth - 40; // Subtract padding
        setContainerWidth(Math.max(600, width)); // Minimum width of 600
      }
    };

    updateWidth();
    window.addEventListener('resize', updateWidth);
    return () => window.removeEventListener('resize', updateWidth);
  }, []);

  useEffect(() => {
    // Handle scope data updates
    const handleScopeData = (data) => {
      setScopes(prev => ({
        ...prev,
        [data.id]: {
          ...prev[data.id],
          ...data,
          lastUpdate: Date.now()
        }
      }));
    };

    // Handle scope list changes
    const handleScopesChanged = (data) => {
      setScopeList(data.scopes || []);

      // Remove scopes that no longer exist
      setScopes(prev => {
        try {
          const activeIds = new Set(data.scopes.map(s => s.id));
          const newScopes = {};
          for (const [id, scope] of Object.entries(prev)) {
            if (activeIds.has(id)) {
              newScopes[id] = scope;
            }
          }
          return newScopes;
        } catch (error) {
          console.error("Error handling scopes changed:", error);
          return prev; // Fallback to previous state on error
        }
      });
    };

    // Subscribe to events
    const unsubscribeData = Events.On('scope.data', (evt) => handleScopeData(evt.data[0]));
    const unsubscribeList = Events.On('scopes-changed', (evt) => handleScopesChanged(evt.data[0]));

    // Cleanup
    return () => {
      unsubscribeData();
      unsubscribeList();
    };
  }, []);

  // Clean up stale scopes (not updated in last 5 seconds)
  useEffect(() => {
    const interval = setInterval(() => {
      const now = Date.now();
      const staleTimeout = 5000; // 5 seconds
      
      setScopes(prev => {
        const newScopes = {};
        for (const [id, scope] of Object.entries(prev)) {
          if (now - scope.lastUpdate < staleTimeout) {
            newScopes[id] = scope;
          }
        }
        return newScopes;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, []);

  const activeScopeIds = scopeList.map(s => s.id);
  const scopesToDisplay = activeScopeIds
    .map(id => scopes[id])
    .filter(scope => scope && scope.samples);

  if (scopesToDisplay.length === 0) {
    return (
      <div className={styles.container}>
        <div className={styles.empty}>
          <p>No active scopes</p>
          <p className={styles.hint}>
            Use the <code>scope</code> function in your code to visualize signals
          </p>
          <pre className={styles.example}>
{`(play (scope (sin 440)))`}
          </pre>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.container} ref={containerRef}>
      <div 
        className={styles.grid}
        style={{
          gridTemplateColumns: '1fr',
          gridAutoRows: 'min-content'
        }}
      >
        {scopesToDisplay.map(scope => (
          <ScopeDisplay
            key={scope.id}
            id={scope.id}
            samples={scope.samples}
            sampleRate={scope.sampleRate}
            name={scope.name}
            width={containerWidth}
            height={250}
          />
        ))}
      </div>
    </div>
  );
};

export default ScopeView;
