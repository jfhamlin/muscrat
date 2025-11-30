import { useState, useEffect, useRef } from 'react';

const SuperColliderTest = () => {
  const [supersonicInstance, setSupersonicInstance] = useState(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const [isPlaying, setIsPlaying] = useState(false);
  const [error, setError] = useState(null);
  const [log, setLog] = useState([]);
  const synthNodeId = useRef(-1);

  const addLog = (message) => {
    setLog((prev) => [...prev, `${new Date().toLocaleTimeString()}: ${message}`]);
  };

  useEffect(() => {
    const initializeSuperSonic = async () => {
      try {
        addLog('Loading SuperSonic library...');

        // Import SuperSonic from the installed npm package
        const { SuperSonic } = await import('supersonic-scsynth');

        addLog('Creating SuperSonic instance...');
        const baseURL = '/supersonic';
        const supersonic = new SuperSonic({
          workerBaseURL: `${baseURL}/workers/`,
          wasmBaseURL: `${baseURL}/wasm/`,
          // We'll skip synthdefs and samples for basic sine wave test
          synthdefBaseURL: `${baseURL}/synthdefs/`,
          sampleBaseURL: `${baseURL}/samples/`,
        });

        addLog('Initializing SuperSonic...');
        await supersonic.init();

        setSupersonicInstance(supersonic);
        setIsInitialized(true);
        addLog('SuperSonic initialized successfully!');
      } catch (err) {
        console.error('Failed to initialize SuperSonic:', err);
        setError(err.message);
        addLog(`ERROR: ${err.message}`);
      }
    };

    initializeSuperSonic();

    return () => {
      if (supersonicInstance) {
        addLog('Cleaning up SuperSonic...');
        supersonicInstance.quit();
      }
    };
  }, []);

  const playSineWave = async () => {
    if (!supersonicInstance) {
      addLog('ERROR: SuperSonic not initialized');
      return;
    }

    try {
      addLog('Playing sine wave at 440 Hz...');

      // Create a sine oscillator using SuperCollider's /s_new command
      // SinOsc is a built-in UGen in SuperCollider
      const nodeId = 1000 + Math.floor(Math.random() * 1000);
      synthNodeId.current = nodeId;

      // /s_new synthDefName nodeID addAction targetID [args]
      // We'll try to use a simple sine synthdef
      // For a basic test, we can try sending OSC commands directly
      supersonicInstance.send('/s_new', 'default', nodeId, 0, 0,
        'freq', 440,
        'amp', 0.3
      );

      setIsPlaying(true);
      addLog(`Started synth node ${nodeId}`);
    } catch (err) {
      console.error('Failed to play sine wave:', err);
      addLog(`ERROR: ${err.message}`);
    }
  };

  const stopSound = () => {
    if (!supersonicInstance) {
      addLog('ERROR: SuperSonic not initialized');
      return;
    }

    try {
      addLog('Stopping sound...');

      if (synthNodeId.current !== -1) {
        // /n_free nodeID - free the synth node
        supersonicInstance.send('/n_free', synthNodeId.current);
        addLog(`Freed synth node ${synthNodeId.current}`);
        synthNodeId.current = -1;
      }

      // Also send a panic message to stop all nodes
      supersonicInstance.send('/g_freeAll', 0);

      setIsPlaying(false);
      addLog('Sound stopped');
    } catch (err) {
      console.error('Failed to stop sound:', err);
      addLog(`ERROR: ${err.message}`);
    }
  };

  const testOscCommand = () => {
    if (!supersonicInstance) {
      addLog('ERROR: SuperSonic not initialized');
      return;
    }

    try {
      // Send a status request
      addLog('Sending /status command...');
      supersonicInstance.send('/status');
      addLog('Status command sent');
    } catch (err) {
      console.error('Failed to send status:', err);
      addLog(`ERROR: ${err.message}`);
    }
  };

  return (
    <div className="p-6 h-full overflow-auto bg-background-primary text-white">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-3xl font-bold mb-4">SuperCollider (SuperSonic) Test Page</h1>

        <div className="mb-6">
          <h2 className="text-xl font-semibold mb-2">Status</h2>
          <div className="bg-gray-800 p-4 rounded">
            <p>Initialized: <span className={isInitialized ? 'text-green-400' : 'text-red-400'}>
              {isInitialized ? 'Yes' : 'No'}
            </span></p>
            <p>Playing: <span className={isPlaying ? 'text-green-400' : 'text-gray-400'}>
              {isPlaying ? 'Yes' : 'No'}
            </span></p>
            {error && <p className="text-red-400 mt-2">Error: {error}</p>}
          </div>
        </div>

        <div className="mb-6">
          <h2 className="text-xl font-semibold mb-2">Controls</h2>
          <div className="flex gap-3">
            <button
              onClick={playSineWave}
              disabled={!isInitialized || isPlaying}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 disabled:bg-gray-600 disabled:cursor-not-allowed rounded"
            >
              Play 440Hz Sine Wave
            </button>
            <button
              onClick={stopSound}
              disabled={!isInitialized || !isPlaying}
              className="px-4 py-2 bg-red-600 hover:bg-red-700 disabled:bg-gray-600 disabled:cursor-not-allowed rounded"
            >
              Stop Sound
            </button>
            <button
              onClick={testOscCommand}
              disabled={!isInitialized}
              className="px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-600 disabled:cursor-not-allowed rounded"
            >
              Test /status Command
            </button>
          </div>
        </div>

        <div className="mb-6">
          <h2 className="text-xl font-semibold mb-2">About</h2>
          <div className="bg-gray-800 p-4 rounded">
            <p className="mb-2">
              This page tests the SuperSonic library, which is a WebAssembly port of
              SuperCollider's scsynth audio synthesis engine.
            </p>
            <p className="mb-2">
              The "Play" button sends a <code className="bg-gray-700 px-1 rounded">/s_new</code> OSC
              command to create a sine wave oscillator at 440 Hz (A4 note).
            </p>
            <p>
              SuperCollider is a powerful audio synthesis platform used for algorithmic composition,
              live coding, and sound design.
            </p>
          </div>
        </div>

        <div className="mb-6">
          <h2 className="text-xl font-semibold mb-2">Log</h2>
          <div className="bg-gray-900 p-4 rounded h-64 overflow-auto font-mono text-sm">
            {log.length === 0 ? (
              <p className="text-gray-500">No logs yet...</p>
            ) : (
              log.map((entry, idx) => (
                <div key={idx} className="mb-1">
                  {entry}
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SuperColliderTest;
