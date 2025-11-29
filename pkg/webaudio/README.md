# WebAudio Package

The WebAudio package enables Muscrat to broadcast audio parameters to web clients (mobile devices, tablets, etc.) over WebSocket, allowing for remote synthesis and visualization.

## Architecture

```
┌─────────────────┐
│  Muscrat (Go)   │
│                 │
│  WebAudioNode   │ ← Samples inputs at low frequency (default 20Hz)
│   (UGen)        │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  HTTP Server    │
│  + WebSocket    │
│                 │
│  (localhost)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  ngrok Tunnel   │ ← Exposes server to internet
│                 │
│ https://xyz...  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Mobile Browser │
│                 │
│  Web Audio API  │ ← Receives params, generates audio
│  + Hydra        │ ← Receives params, generates visuals
└─────────────────┘
```

## Components

### 1. Server (`server.go`)

The HTTP/WebSocket server that:
- Serves the client HTML page via embedded `index.html`
- Manages WebSocket connections for multiple clients
- Broadcasts parameter updates to all connected clients
- Integrates with ngrok to create a public tunnel

**Key Methods:**
- `NewServer(port int)` - Creates a new server instance
- `Start()` - Starts HTTP server and ngrok tunnel
- `Stop()` - Gracefully shuts down server and closes connections
- `Broadcast(msg Message)` - Sends a message to all connected clients
- `GetURL()` - Returns the public ngrok URL
- `GetClientCount()` - Returns number of connected clients

### 2. WebAudioNode (`ugen/webaudio.go`)

A UGen that samples input signals and broadcasts them to connected devices.

**Key Features:**
- Implements `UGen` interface with `Gen()` method
- Implements `Starter` and `Stopper` for lifecycle management
- Samples inputs at configurable rate (default 20Hz)
- Tracks parameter changes to avoid redundant broadcasts
- Returns zeros for audio output (it's a control node, not an audio generator)

**Constructor:**
```go
NewWebAudioNode(port int, updateHz float64, paramNames []string)
```

### 3. Web Client (`static/index.html`)

A mobile-optimized web page that:
- Connects to the server via WebSocket
- Receives parameter updates in real-time
- Synthesizes audio using Web Audio API
- Visualizes parameters using Hydra
- Auto-reconnects on disconnect

**Web Audio Graph:**
```
Oscillator → LowpassFilter → Gain → Destination
```

**Controllable Parameters:**
- `freq` - Oscillator frequency (default: 440 Hz)
- `amp` - Output amplitude (default: 0.3)
- `cutoff` - Lowpass filter cutoff frequency (default: 2000 Hz)
- `resonance` - Filter resonance/Q (default: 1)

## Message Protocol

JSON messages sent over WebSocket:

### Server → Client

**Parameter Update:**
```json
{
  "type": "param",
  "params": {
    "freq": 440.0,
    "amp": 0.5,
    "cutoff": 1000.0
  }
}
```

**Synth Control:**
```json
{
  "type": "synth",
  "action": "start"  // or "stop"
}
```

**Connection Confirmation:**
```json
{
  "type": "connected",
  "action": "ready"
}
```

### Client → Server

**Heartbeat:**
```json
{
  "type": "ping"
}
```

**Heartbeat Response:**
```json
{
  "type": "pong"
}
```

## Glojure API

### `web-audio`

Creates a WebAudio node that broadcasts parameters to connected clients.

**Signature:**
```clojure
(web-audio params & {:keys [port update-hz]})
```

**Parameters:**
- `params` - Map of parameter names to UGen values
- `port` - (optional) HTTP server port (default: 8765)
- `update-hz` - (optional) Update frequency in Hz (default: 20)

**Example:**
```clojure
(web-audio {:freq (sin 0.5 :mul 100 :add 440)
            :amp (knob "amp" 0.5 0 1)
            :cutoff (* 1000 (knob "cutoff" 0.5 0 1))}
           :port 8080
           :update-hz 30)
```

## Usage Flow

1. **Start Muscrat** with a WebAudio node in your code
2. **Server starts** on specified port and creates ngrok tunnel
3. **ngrok URL printed** to console (e.g., `https://abc123.ngrok.io`)
4. **Open URL on phone** in any modern browser
5. **Tap "Start Audio"** to initialize Web Audio context
6. **Parameters stream** from Muscrat to phone at configured rate
7. **Phone synthesizes audio** and visualizes with Hydra

## Requirements

- **ngrok** must be installed and available in PATH
  - Install: `brew install ngrok` (macOS) or download from https://ngrok.com
- Modern web browser with Web Audio API support (all recent mobile browsers)
- Internet connection for ngrok tunnel

## Performance Notes

- **Low bandwidth:** Parameters sent at low frequency (default 20Hz) to minimize data usage
- **Change detection:** Only sends updates when values change significantly (>0.001)
- **Multiple devices:** Supports unlimited simultaneous connections
- **Auto-reconnect:** Client automatically reconnects if connection drops

## Security Considerations

- ngrok URLs are publicly accessible (anyone with URL can connect)
- No authentication implemented (suitable for temporary, personal use)
- For production use, consider:
  - Authentication/authorization
  - Private ngrok tunnels (requires paid ngrok account)
  - Direct LAN connections without ngrok
  - HTTPS with proper certificates

## Future Enhancements

- [ ] Device-specific addressing (send different params to different devices)
- [ ] Bidirectional control (send events from phone to Muscrat)
- [ ] Touch/gesture input from mobile devices
- [ ] Pre-defined synth templates on client
- [ ] Recording/playback of parameter streams
- [ ] Custom Hydra code upload from Muscrat
- [ ] Multiple synth engines on client (polyphony)
- [ ] WebRTC for lower latency
